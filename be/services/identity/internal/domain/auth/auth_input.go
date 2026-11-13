package auth

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}
