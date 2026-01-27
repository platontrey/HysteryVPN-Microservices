package errors

import "fmt"

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	Message string
}

func (e AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Message)
}

// AuthorizationError represents authorization failures
type AuthorizationError struct {
	Message string
}

func (e AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error: %s", e.Message)
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// ConflictError represents resource conflict errors
type ConflictError struct {
	Resource string
	Message  string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict in %s: %s", e.Resource, e.Message)
}

// InternalError represents internal server errors
type InternalError struct {
	Message string
}

func (e InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Message)
}
