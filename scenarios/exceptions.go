package scenarios

type StepValidationException struct {
	message string
}

func (e *StepValidationException) Error() string {
	return e.message
}

type StepNotFoundException struct {
	message string
}

func (e *StepNotFoundException) Error() string {
	return e.message
}
