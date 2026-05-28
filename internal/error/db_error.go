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
	DBErrInit              = errors.New("db init failed")
	DBErrQuery             = errors.New("failed to select")
	DBErrInsert            = errors.New("failed to insert")
	DBErrUpdate            = errors.New("failed to insert")
	DBErrForeignConstraint = errors.New("failed on foreign key constraint")
	DBErrBadArgument       = errors.New("bad argument passed")
	DBErrSerialization     = errors.New("failed to serialize output")
	DBErrFatal             = errors.New("fatal internal error")
	DBErrNotFound          = errors.New("found no rows")
)

func (dbErr DBError) Error() string {
	return fmt.Sprintf("[%s]: %s on table: %s, action: %s, values: %+q", dbErr.Context, dbErr.Err, dbErr.Table, dbErr.Action, dbErr.Values)
}
