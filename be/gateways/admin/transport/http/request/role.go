package request

type RoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
type ReplaceRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}
