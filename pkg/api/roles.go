package api

import (
	"encoding/json"
	"net/http"
	"time"

	eeDStore "github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
)

type RolesController struct {
	db     *database.DB
	eStore eeDStore.Store
}

func NewRolesController(db *database.DB, eStore eeDStore.Store) *RolesController {
	return &RolesController{
		db:     db,
		eStore: eStore,
	}
}

func (c *RolesController) MountRouter(r chi.Router) {
	r.Get("/{roleName}", c.get)
	r.Delete("/{roleName}", c.delete)
	r.Put("/{roleName}", c.update)

	r.Get("/", c.list)
	r.Post("/", c.create)
}

func (c *RolesController) get(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	roleName := chi.URLParam(r, "roleName")

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	// Fetch one
	role, err := c.eStore.With(db.Conn()).Roles().Get(r.Context(), ns.Name, roleName)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, convertRole(role))
}

func (c *RolesController) delete(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	roleName := chi.URLParam(r, "roleName")

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	err = c.eStore.With(db.Conn()).Roles().Delete(r.Context(), ns.Name, roleName)
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

func (c *RolesController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	// Parse request.
	req := struct {
		Name        string               `json:"name"`
		Description string               `json:"description"`
		OidcGroups  eeDStore.OidcGroups  `json:"oidcGroups"`
		Permissions eeDStore.Permissions `json:"permissions"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	err = req.OidcGroups.Validate()
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field oidcGroup has invalid format",
			Validation: map[string]string{
				"permissions": err.Error(),
			},
		})

		return
	}

	// Create role.
	role, err := c.eStore.With(db.Conn()).Roles().Create(r.Context(), &eeDStore.Role{
		Name:        req.Name,
		Namespace:   ns.Name,
		Description: req.Description,
		OidcGroups:  req.OidcGroups,
		Permissions: req.Permissions,
	})
	if err != nil {
		writeDataStoreError(w, err)

		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, convertRole(role))
}

func (c *RolesController) update(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	roleName := chi.URLParam(r, "roleName")

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	// Parse request.
	req := struct {
		Name        string               `json:"name"`
		Description string               `json:"description"`
		OidcGroups  eeDStore.OidcGroups  `json:"oidcGroups"`
		Permissions eeDStore.Permissions `json:"permissions"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	err = req.OidcGroups.Validate()
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field permissions has invalid format",
			Validation: map[string]string{
				"permissions": err.Error(),
			},
		})

		return
	}

	// Update role.
	role, err := c.eStore.With(db.Conn()).Roles().Update(r.Context(), ns.Name, roleName, &eeDStore.Role{
		Name:        req.Name,
		Description: req.Description,
		OidcGroups:  req.OidcGroups,
		Permissions: req.Permissions,
	})
	if err != nil {
		writeDataStoreError(w, err)

		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, convertRole(role))
}

func (c *RolesController) list(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := c.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	list, err := c.eStore.With(db.Conn()).Roles().List(r.Context(), ns.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	res := make([]any, len(list))
	for i := range list {
		res[i] = convertRole(list[i])
	}

	writeJSON(w, res)
}

func convertRole(v *eeDStore.Role) any {
	type secretForAPI struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OidcGroups  any    `json:"oidcGroups"`
		Permissions any    `json:"permissions"`

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

	res := &secretForAPI{
		Name:        v.Name,
		Description: v.Description,
		OidcGroups:  v.OidcGroups,
		Permissions: permissions,

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}

	return res
}
