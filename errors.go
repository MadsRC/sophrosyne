package sophrosyne

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
)

type UnreachableCodeError struct {
	stack []byte
}

func NewUnreachableCodeError() error {
	stack := debug.Stack()
	return &UnreachableCodeError{
		stack: stack,
	}
}

func (e UnreachableCodeError) Error() string {
	return fmt.Sprintf("unreachable code encountered - this is a bug.\nStack:\n%s", e.stack)
}

func (e UnreachableCodeError) LogValue() slog.Value {
	return slog.GroupValue(slog.String("stack", string(e.stack)))
}

type PanicError struct {
	reason string
	stack  []byte
}

func (e PanicError) Error() string {
	return fmt.Sprintf("panic encountered.\nReason: %s\nStack:\n%s", e.reason, e.stack)
}

func (e PanicError) LogValue() slog.Value {
	return slog.GroupValue(slog.String("reason", e.reason), slog.String("stack", string(e.stack)))
}

var ErrNotFound = errors.New("not found")

type ConstraintViolationError struct {
	UnderlyingError error
	code            string
	Detail          string
	TableName       string
	ConstraintName  string
}

type DatastoreError interface {
	error
	Code() string
}

func NewConstraintViolationError(err error, code, detail, tableName, constraintName string) error {
	return &ConstraintViolationError{
		UnderlyingError: err,
		code:            code,
		Detail:          detail,
		TableName:       tableName,
		ConstraintName:  constraintName,
	}
}

func (e ConstraintViolationError) Error() string {
	return fmt.Sprintf("violation of constraint '%s' in table '%s' - code '%s'. Detail: %s", e.ConstraintName, e.TableName, e.code, e.Detail)
}

func (e ConstraintViolationError) Code() string {
	return e.code
}
