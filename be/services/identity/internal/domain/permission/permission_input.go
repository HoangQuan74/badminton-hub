package permission

type Input struct {
	Module      string `validate:"required,max=100"`
	Path        string `validate:"required,max=300"`
	Action      string `validate:"required,max=6"`
	Description string `validate:"max=500"`
}
