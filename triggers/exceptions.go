package triggers

type TriggerValidationException struct {
	message string
}

func (e *TriggerValidationException) Error() string {
	return e.message
}

type TriggerNotFoundException struct {
	message string
}

func (e *TriggerNotFoundException) Error() string {
	return e.message
}
