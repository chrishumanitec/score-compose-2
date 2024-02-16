package cmd

import (
	"errors"
	"fmt"
)

var ErrUsage = errors.New("usage error")

func NewValidationError(msg string, v ...interface{}) error {
	return &CmdError{
		message: fmt.Sprintf(msg, v...),
		err:     ErrUsage,
	}
}

type CmdError struct {
	message string
	err     error
}

func (c *CmdError) Error() string {
	if c.err != nil && c.err != ErrUsage {
		return fmt.Sprintf("%s: %v", c.message, c.err)
	}
	return c.message
}

func (c *CmdError) Unwrap() error {
	return c.err
}

func (e *CmdError) Is(target error) bool {
	t, ok := target.(*CmdError)
	if !ok {
		return false
	}
	return e.message == t.message && errors.Is(e.err, t.err)
}

func NewCmdError(msg string, v ...interface{}) error {
	return &CmdError{
		message: fmt.Sprintf(msg, v...),
	}
}

func NewCmdErrorWrap(err error, msg string, v ...interface{}) error {
	return &CmdError{
		message: fmt.Sprintf(msg, v...),
		err:     err,
	}
}
