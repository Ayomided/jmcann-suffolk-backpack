package error

import (
	"errors"
	"fmt"
)

type ServerError struct {
	Handler string
	Err     error
}

var (
	ServerErrFailed          = errors.New("failed")
	ServerErrBadRequest      = errors.New("bad request")
	ServerErrUnauthorized    = errors.New("unauthorized")
	ServerErrForbidden       = errors.New("forbidden")
	ServerErrNotFound        = errors.New("not found")
	ServerErrInvalidForm     = errors.New("invalid form data")
	ServerErrInvalidID       = errors.New("invalid id")
	ServerErrTemplateMissing = errors.New("template does not exist")
	ServerErrTemplateRender  = errors.New("failed to render template")
	ServerErrInternal        = errors.New("internal server error")
)

func (se ServerError) Error() string {
	return fmt.Sprintf("[%s]: %s", se.Handler, se.Err)
}
