package main

import (
	"fmt"
)

// ErrorContext provides structured error information
type ErrorContext struct {
	Operation string
	Context   map[string]interface{}
	Err       error
}

// Error implements the error interface
func (e *ErrorContext) Error() string {
	return fmt.Sprintf("%s: %v (context: %v)", e.Operation, e.Err, e.Context)
}

// Unwrap returns the underlying error
func (e *ErrorContext) Unwrap() error {
	return e.Err
}

// NewError creates a new error with context
func NewError(operation string, err error, context map[string]interface{}) *ErrorContext {
	return &ErrorContext{
		Operation: operation,
		Err:       err,
		Context:   context,
	}
}

// WrapError wraps an error with an operation description
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// WrapErrorf wraps an error with a formatted operation description
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	operation := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", operation, err)
}
