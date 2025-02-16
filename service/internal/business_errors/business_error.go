package business_errors

import "fmt"

type BusinessError struct {
	Code string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("business error (%s)", e.Code)
}

func New(code string) *BusinessError {
	return &BusinessError{code}
}
