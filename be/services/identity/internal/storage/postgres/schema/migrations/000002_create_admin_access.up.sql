CREATE TABLE identity.roles (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX roles_name_unique_idx ON identity.roles (LOWER(name));

CREATE TABLE identity.permissions (
    id UUID PRIMARY KEY,
    module TEXT NOT NULL,
    path TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('CREATE', 'READ', 'UPDATE', 'DELETE')),
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX permissions_module_path_action_unique_idx
    ON identity.permissions (LOWER(module), LOWER(path), action);

CREATE TABLE identity.role_permissions (
    role_id UUID NOT NULL REFERENCES identity.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES identity.permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

ALTER TABLE identity.users
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN role_id UUID REFERENCES identity.roles(id) ON DELETE SET NULL;

CREATE INDEX users_role_id_idx ON identity.users (role_id);
