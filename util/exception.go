package util

import "fmt"

type ParamValidationException struct {
	message string
}

func CreateParamValidationException(param string, err error) *ParamValidationException {
	return &ParamValidationException{message: fmt.Sprintf("Parameter %s is not valid: %v", param, err)}
}

func (e *ParamValidationException) Error() string {
	return e.message
}
