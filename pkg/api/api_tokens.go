package api

import (
	"encoding/json"
	"net/http"
	"time"

	isoDuration "github.com/ChannelMeter/iso8601duration"
	eeDStore "github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

//nolint:revive
type APITokensController struct {
	db     *database.DB
	eStore eeDStore.Store
}

func NewAPITokensController(db *database.DB, eStore eeDStore.Store) *APITokensController {
	return &APITokensController{
		db:     db,
		eStore: eStore,
	}
}

func (c *APITokensController) MountRouter(r chi.Router) {
	r.Get("/{apiTokenName}", c.get)
	r.Delete("/{apiTokenName}", c.delete)

	r.Get("/", c.list)
	r.Post("/", c.create)
}

func (c *APITokensController) get(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	apiTokenName := chi.URLParam(r, "apiTokenName")

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	// Fetch one
	apiToken, err := c.eStore.With(db.Conn()).APITokens().Get(r.Context(), ns.Name, apiTokenName)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, convertAPIToken(apiToken))
}

func (c *APITokensController) delete(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	apiTokenName := chi.URLParam(r, "apiTokenName")

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	err = c.eStore.With(db.Conn()).APITokens().Delete(r.Context(), ns.Name, apiTokenName)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeOk(w)
}

func (c *APITokensController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	// Parse request.
	req := struct {
		Name            string               `json:"name"`
		Description     string               `json:"description"`
		Permissions     eeDStore.Permissions `json:"permissions"`
		DurationISO8601 string               `json:"duration"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	// Parse the ISO 8601 in duration field.
	duration, err := isoDuration.FromString(req.DurationISO8601)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "request data has invalid fields",
			Validation: map[string]string{
				"duration": "invalid iso8601 duration format",
			},
		})

		return
	}

	secret := uuid.New()
	hash := eeDStore.HashTokenID(secret)

	// Create apiToken.
	apiToken, err := c.eStore.With(db.Conn()).APITokens().Create(r.Context(), &eeDStore.APIToken{
		Name:        req.Name,
		Namespace:   ns.Name,
		Description: req.Description,
		Hash:        hash,
		Permissions: req.Permissions,
	}, int(duration.ToDuration().Seconds()))
	if err != nil {
		writeDataStoreError(w, err)

		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	type res struct {
		APIToken any    `json:"apiToken"`
		Secret   string `json:"secret"`
	}

	writeJSON(w, &res{
		APIToken: convertAPIToken(apiToken),
		Secret:   secret.String(),
	})
}

func (c *APITokensController) list(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	list, err := c.eStore.With(db.Conn()).APITokens().List(r.Context(), ns.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	res := make([]any, len(list))
	for i := range list {
		res[i] = convertAPIToken(list[i])
	}

	writeJSON(w, res)
}

func convertAPIToken(v *eeDStore.APIToken) any {
	type apiTokenForAPI struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Prefix      string    `json:"prefix"`
		Permissions any       `json:"permissions"`
		ExpiredAt   time.Time `json:"expiredAt"`
		IsExpired   bool      `json:"isExpired"`

		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	type permission struct {
		Topic  string `json:"topic"`
		Method string `json:"method"`
	}

	permissions := make([]permission, len(v.Permissions))
	for i := range permissions {
		permissions[i] = permission{
			Topic:  v.Permissions[i].Topic,
			Method: v.Permissions[i].Method,
		}
	}
	if v.Permissions == nil {
		permissions = nil
	}

	res := &apiTokenForAPI{
		Name:        v.Name,
		Description: v.Description,
		Prefix:      v.Hash.String()[0:8],
		Permissions: permissions,
		ExpiredAt:   v.ExpiredAt,
		IsExpired:   v.IsExpired,

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}

	return res
}
