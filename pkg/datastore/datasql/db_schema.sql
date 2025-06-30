CREATE TABLE IF NOT EXISTS "ee_roles" (
    "name" text NOT NULL,
    "namespace" text NOT NULL,
    "description" text NOT NULL,
    "oidc_groups" text NOT NULL,
    "permissions" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("name", "namespace"),
    CONSTRAINT "fk_namespaces_ee_roles"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "ee_api_tokens" (
    "name" text NOT NULL,
    "namespace" text NOT NULL,
    "description" text NOT NULL,
    "hash" uuid NOT NULL,
    "permissions" text NOT NULL,
    "expired_at" timestamptz NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("name", "namespace"),
    UNIQUE ("hash"),
    CONSTRAINT "fk_namespaces_ee_api_tokens"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);
