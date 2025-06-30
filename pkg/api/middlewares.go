package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	eeDStore "github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type Middlewares struct {
	db     *database.DB
	config *core.Config
	eStore eeDStore.Store
	lru    *expirable.LRU[string, string]
}

func NewMiddlewares(db *database.DB, config *core.Config, eStore eeDStore.Store, lru *expirable.LRU[string, string]) *Middlewares {
	return &Middlewares{
		db:     db,
		config: config,
		eStore: eStore,
		lru:    lru,
	}
}

func (c *Middlewares) CheckOidc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			next.ServeHTTP(w, r)
			return
		}
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		authHeader = strings.TrimPrefix(authHeader, "bearer ")
		oidcGroups, ok := c.lru.Get(authHeader)
		if ok {
			r.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
			r.Header.Set("X-Oidc-Groups", oidcGroups)
			next.ServeHTTP(w, r)

			return
		}

		if os.Getenv("DIREKTIV_OIDC_DEV") == "true" {
			c.config.OidcIssuerUrl = "http://127.0.0.1:9090/dex"
			c.config.OidcClientID = "direktiv"
		}

		// Use the original authHeader for claims extraction
		oidcGroups, err := extractOidcGroupsFromToken(r.Context(), c.config.OidcIssuerUrl, c.config.OidcClientID, authHeader)
		if err != nil {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "couldn't parse claims from jwt token",
			})

			return
		}
		if oidcGroups == "" {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "empty oidc groups in claims",
			})

			return
		}

		c.lru.Add(authHeader, oidcGroups)
		r.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
		r.Header.Set("X-Oidc-Groups", oidcGroups)
		next.ServeHTTP(w, r)
	})
}

func (c *Middlewares) CheckAPIToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiTokenStr := r.Header.Get("Direktiv-Api-Token")
		if apiTokenStr == "" {
			next.ServeHTTP(w, r)
			return
		}
		apiToken, err := uuid.Parse(apiTokenStr)
		if err != nil {
			writeError(w, &Error{
				Code:    "access_token_invalid",
				Message: "api token invalid format",
			})

			return
		}
		permissions, ok := c.lru.Get(apiTokenStr)
		if ok {
			r.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
			r.Header.Set("X-Permissions", permissions)
			next.ServeHTTP(w, r)

			return
		}

		t, err := c.eStore.With(c.db.Conn()).APITokens().GetByHash(r.Context(), eeDStore.HashTokenID(apiToken))
		if errors.Is(err, eeDStore.ErrNotFound) {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "api token is denied",
			})

			return
		}
		if err != nil {
			writeInternalError(w, err)
			return
		}
		if t.IsExpired {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "api token is expired",
			})

			return
		}
		c.lru.Add(apiTokenStr, t.Permissions.String())
		r.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
		r.Header.Set("X-Permissions", t.Permissions.String())
		next.ServeHTTP(w, r)
	})
}

//nolint:gocognit,goconst
func (c *Middlewares) CheckAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:gosec
		apiKeyHeader := "Direktiv-Api-Key"
		apiKey := os.Getenv("DIREKTIV_API_KEY")
		if apiKey == "" {
			next.ServeHTTP(w, r)

			return
		}
		if r.Header.Get(apiKeyHeader) == "" {
			writeError(w, &Error{
				Code:    "access_token_missing",
				Message: "missing api key",
			})

			return
		}
		if apiKey != r.Header.Get(apiKeyHeader) {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "invalid api key",
			})

			return
		}

		// this is direct access with api key
		if r.Header.Get("X-Oidc-Groups") == "" && r.Header.Get("X-Permissions") == "" {
			next.ServeHTTP(w, r)

			return
		}

		reqGroupsStr := r.Header.Get("X-Oidc-Groups")

		reqGroups := strings.Split(reqGroupsStr, ",")

		// Admin group has like a root access.
		if slices.Contains(reqGroups, os.Getenv("DIREKTIV_OIDC_ADMIN_GROUP")) {
			next.ServeHTTP(w, r)

			return
		}

		reqNamespace, reqTopic := extractNamespaceAndTopic(r.URL.Path)

		// None admins cannot create namespaces.
		if !slices.Contains(reqGroups, os.Getenv("DIREKTIV_OIDC_ADMIN_GROUP")) &&
			r.Method == http.MethodPost &&
			reqTopic == "namespaces" &&
			reqNamespace == "" {
			writeError(w, &Error{
				Code:    "access_token_denied",
				Message: "only admins can create namespaces",
			})

			return
		}

		roles, err := c.eStore.With(c.db.Conn()).Roles().ListAll(r.Context())
		if err != nil {
			writeInternalError(w, err)
		}

		var permissions eeDStore.Permissions
		if r.Header.Get("X-Permissions") != "" {
			_ = permissions.Scan(r.Header.Get("X-Permissions"))
		}

		for _, group := range reqGroups {
			for _, role := range roles {
				if slices.Contains(role.OidcGroups, group) {
					permissions = append(permissions, role.Permissions...)
				}
			}
		}

		allowedNamespaces := ","
		for _, permission := range permissions {
			allowedNamespaces += permission.Namespace + ","
		}

		for _, permission := range permissions {
			if permission.Namespace != reqNamespace && reqNamespace != "" {
				continue
			}
			if permission.Topic != reqTopic {
				continue
			}
			if permission.Method == "read" {
				permission.Method = "GET"
			}
			if permission.Method == "manage" || permission.Method == r.Method {
				req := r.WithContext(context.WithValue(r.Context(), "allowedNamespaces", allowedNamespaces))
				next.ServeHTTP(w, req)

				return
			}
		}

		writeError(w, &Error{
			Code:    "access_token_denied",
			Message: "not enough permissions",
		})
	})
}

// nolint
func extractOidcGroupsFromToken(ctx context.Context, oidcIssuerURL, oidcClientID, oidcToken string) (string, error) {
	if os.Getenv("DIREKTIV_OIDC_DEV") == "true" {
		return "admin,g1,g2", nil
	}

	if os.Getenv("DIREKTIV_OIDC_SKIP_TLS_VERIFY") == "true" {
		insecureClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // ⚠️ This disables SSL verification!
				},
			},
		}
		// Create an OIDC provider using the insecure client
		ctx = oidc.ClientContext(context.Background(), insecureClient)
	}

	provider, err := oidc.NewProvider(ctx, oidcIssuerURL)
	if err != nil {
		return "", fmt.Errorf("error creating oidc provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oidcClientID})
	oidcTokenObject, err := verifier.Verify(ctx, oidcToken)
	if err != nil {
		return "", fmt.Errorf("error verifying token: %w", err)
	}

	claims := make(map[string]interface{})
	if err := oidcTokenObject.Claims(&claims); err != nil {
		return "", fmt.Errorf("error parsing token claims: %w", err)
	}
	groups := parseOIDCGroups(claims)

	return strings.Join(groups, ","), nil
}

// nolint
func parseOIDCGroups(claims map[string]interface{}) []string {
	var parsedGroups []string
	if groups, ok := claims["user_groups"].([]interface{}); ok {
		for _, g := range groups {
			if group, ok := g.(string); ok {
				parsedGroups = append(parsedGroups, group)
			}
		}
	}

	return parsedGroups
}

// nolint
func getClaim(claims map[string]interface{}, key string, fallback string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}

	return fallback
}

//nolint:goconst
func extractNamespaceAndTopic(pathString string) (string, string) {
	pathString = "/" + pathString + "/"
	pathString = path.Clean(pathString)

	pathString = strings.TrimPrefix(pathString, "/api/v1")
	pathString = strings.TrimPrefix(pathString, "/api/v2")
	pathString = strings.TrimPrefix(pathString, "/api/v3")
	pathString = strings.TrimPrefix(pathString, "/")

	parts := strings.Split(pathString, "/")
	if len(parts) < 1 {
		return "", ""
	}
	if parts[0] != "namespaces" {
		return "", parts[0]
	}
	if len(parts) < 2 {
		return "", "namespaces"
	}
	if len(parts) < 3 {
		return parts[1], "namespaces"
	}

	return parts[1], parts[2]
}
