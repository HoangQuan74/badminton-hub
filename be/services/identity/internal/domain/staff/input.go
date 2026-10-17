package staff

type CreateInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	FullName string `validate:"required"`
	RoleID   string
	IsAdmin  bool
}

type UpdateInput struct {
	ID       string
	Email    string
	FullName string
	Status   string `validate:"required"`
}

type AccessInput struct {
	ID      string
	RoleID  string
	IsAdmin bool
}
