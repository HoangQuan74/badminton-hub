package request

type CreateStaffRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	FullName string `json:"full_name" binding:"required"`
	IsAdmin  bool   `json:"is_admin"`
	RoleID   string `json:"role_id"`
}
type UpdateStaffRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Status   string `json:"status" binding:"required"`
}
type UpdateStaffAccessRequest struct {
	IsAdmin bool   `json:"is_admin"`
	RoleID  string `json:"role_id"`
}
