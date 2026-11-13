package request

type PermissionRequest struct {
	Module      string `json:"module" binding:"required,max=100"`
	Path        string `json:"path" binding:"required,max=300"`
	Action      string `json:"action" binding:"required,max=6"`
	Description string `json:"description" binding:"max=500"`
}
