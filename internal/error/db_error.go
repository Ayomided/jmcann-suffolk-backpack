package error

import (
	"errors"
	"fmt"
)

type DBError struct {
	Context       string
	Values        []string
	Action, Table string
	Err           error
}

var (
	DBErrQuery             error = errors.New("failed to select")
	DBErrInsert            error = errors.New("failed to insert")
	DBErrUpdate            error = errors.New("failed to insert")
	DBErrForeignConstraint error = errors.New("failed on foreign key constraint")
	DBErrBadArgument       error = errors.New("bad argument passed")
	DBErrSerialization     error = errors.New("failed to serialize output")
)

func (dbErr DBError) Error() string {
	return fmt.Sprintf("[%s]: %s on table: %s, action: %s, values: %+q", dbErr.Context, dbErr.Err, dbErr.Table, dbErr.Action, dbErr.Values)
}
