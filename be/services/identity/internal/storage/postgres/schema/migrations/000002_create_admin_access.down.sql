DROP INDEX IF EXISTS identity.users_role_id_idx;
ALTER TABLE identity.users DROP COLUMN IF EXISTS role_id;
ALTER TABLE identity.users DROP COLUMN IF EXISTS is_admin;
DROP TABLE IF EXISTS identity.role_permissions;
DROP INDEX IF EXISTS identity.permissions_module_path_action_unique_idx;
DROP TABLE IF EXISTS identity.permissions;
DROP INDEX IF EXISTS identity.roles_name_unique_idx;
DROP TABLE IF EXISTS identity.roles;
