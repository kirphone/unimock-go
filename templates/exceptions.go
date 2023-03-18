package templates

type TemplateValidationException struct {
	message string
}

func (e *TemplateValidationException) Error() string {
	return e.message
}

type TemplateNotFoundException struct {
	message string
}

func (e *TemplateNotFoundException) Error() string {
	return e.message
}
