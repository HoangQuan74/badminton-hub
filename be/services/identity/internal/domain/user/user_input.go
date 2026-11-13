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

type RegisterInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=72"`
	FullName string `validate:"required,max=120"`
	Phone    string `validate:"max=20"`
}
