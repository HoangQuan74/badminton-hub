package request

type CreateTenantRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Note         string `json:"note"`
}

type UpdateTenantRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}
