package auth

type Error struct {
	Code string
}

func (e *Error) Error() string { return e.Code }

var (
	ErrEmailTaken         = &Error{Code: "auth.email_already_registered"}
	ErrInvalidInput       = &Error{Code: "auth.invalid_registration"}
	ErrInvalidCredentials = &Error{Code: "auth.invalid_credentials"}
	ErrInvalidToken       = &Error{Code: "auth.invalid_token"}
	ErrUserInactive       = &Error{Code: "auth.user_inactive"}
	ErrUserNotFound       = &Error{Code: "auth.user_not_found"}
	ErrWrongPortal        = &Error{Code: "auth.wrong_portal"}
)
