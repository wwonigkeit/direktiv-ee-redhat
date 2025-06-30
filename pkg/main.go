package main

import (
	"fmt"
	"os"
	"time"

	"github.com/direktiv/direktiv/cmd/cli"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/api"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore/datasql"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/license"
	_ "github.com/direktiv/direktiv/direktiv-ee/pkg/plugins/inbound"
	_ "github.com/direktiv/direktiv/direktiv-ee/pkg/plugins/target"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/extensions"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

func main() {
	cli.Run()
}

// nolint:gochecknoinits
func init() {
	extensions.IsEnterprise = true
	extensions.AdditionalSchema = datasql.Schema

	extensions.Initialize = func(db *database.DB, bus *pubsub.Bus, config *core.Config) error {
		apiCtr := api.NewAPITokensController(db, datasql.New())
		rolesCtr := api.NewRolesController(db, datasql.New())
		mwCtr := api.NewMiddlewares(
			db,
			config,
			datasql.New(),
			expirable.NewLRU[string, string](1000, nil, time.Second*30))

		extensions.AdditionalAPIRoutes = map[string]func(r chi.Router){
			"/namespaces/{namespace}/api_tokens": apiCtr.MountRouter,
			"/namespaces/{namespace}/roles":      rolesCtr.MountRouter,
		}
		extensions.CheckOidcMiddleware = mwCtr.CheckOidc
		extensions.CheckAPITokenMiddleware = mwCtr.CheckAPIToken
		extensions.CheckAPIKeyMiddleware = mwCtr.CheckAPIKey

		// Check if license is valid.
		if os.Getenv("DIREKTIV_LICENSE") == "" {
			return fmt.Errorf("missing DIREKTIV_LICENSE environment variable")
		}
		if err := license.VerifyJSON(os.Getenv("DIREKTIV_LICENSE"), license.PublicKey); err != nil {
			return fmt.Errorf("invalid direktiv license: %w", err)
		}

		if os.Getenv("DIREKTIV_OIDC_ISSUER_URL") != "" {
			if os.Getenv("DIREKTIV_OIDC_ADMIN_GROUP") == "" {
				return fmt.Errorf("missing DIREKTIV_OIDC_ADMIN_GROUP environment variable")
			}
			if os.Getenv("DIREKTIV_OIDC_CLIENT_ID") == "" {
				return fmt.Errorf("missing DIREKTIV_OIDC_CLIENT_ID environment variable")
			}
		}

		return nil
	}
}
