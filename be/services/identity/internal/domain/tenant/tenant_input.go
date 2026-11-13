package tenant

type CreateOrganizationInput struct {
	Name           string `validate:"required,max=120"`
	Slug           string `validate:"required,max=100"`
	IdempotencyKey string `validate:"required,uuid4_rfc4122"`
	ContactName    string `validate:"max=120"`
	ContactEmail   string `validate:"omitempty,email,max=254"`
	ContactPhone   string `validate:"max=30"`
	Note           string `validate:"max=1000"`
}

type UpdateOrganizationInput struct {
	Name *string
	Slug *string
}

type RegistrationContactInput struct {
	ContactName  string
	ContactEmail string
	ContactPhone string
	Note         string
}
type ReviewRegistrationInput struct {
	OrganizationID string
	Decision       string
	Note           string
}
