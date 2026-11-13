package exception

type AuthError struct {
	Code string
}

func (e *AuthError) Error() string { return e.Code }

var (
	ErrAuthEmailTaken         = &AuthError{Code: "auth.email_already_registered"}
	ErrAuthInvalidInput       = &AuthError{Code: "auth.invalid_registration"}
	ErrAuthInvalidCredentials = &AuthError{Code: "auth.invalid_credentials"}
	ErrAuthInvalidToken       = &AuthError{Code: "auth.invalid_token"}
	ErrAuthUserInactive       = &AuthError{Code: "auth.user_inactive"}
	ErrAuthUserNotFound       = &AuthError{Code: "auth.user_not_found"}
	ErrAuthWrongPortal        = &AuthError{Code: "auth.wrong_portal"}
)
