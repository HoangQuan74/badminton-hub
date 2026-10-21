package auth

type RegisterInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=72"`
	FullName string `validate:"required,max=120"`
	Phone    string `validate:"max=20"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
	Portal   string
}
