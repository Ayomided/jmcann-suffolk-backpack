package error

import "fmt"

type ServerError struct {
	Handler string
	Err     error
}

func (se ServerError) Error() string {
	return fmt.Sprintf("[%s]: %s", se.Handler, se.Err)
}
