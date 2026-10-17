package user

type CreateInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
	FullName string `validate:"required"`
	Phone    string
}

type UpdateInput struct {
	ID       string
	Email    string `validate:"required,email"`
	FullName string `validate:"required"`
	Phone    string
	Status   string
}
