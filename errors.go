package eml

import (
	"fmt"
	"net/http"
)

const (
	ErrorUnauthorized      sUserErrorMessage = "Access token is not valid"
	ErrorForbidden         sUserErrorMessage = "You do not have permission to access this resource"
	ErrorInternal          sUserErrorMessage = "An internal error has occurred, please try again later"
	ErrorNotImplemented    sUserErrorMessage = "This operation is currently not supported"
	ErrorValidation        sUserErrorMessage = "Input data has failed validation"
	ErrorParsingBody       sUserErrorMessage = "Error parsing body"
)

const (
	ErrorStatus             fUserErrorMessage = "Invalid %s status: %s"
	//ErrorNotFound           fUserErrorMessage = "%s was not found"
	//ErrorNotFoundPlural     fUserErrorMessage = "Some %s were not found: %s"
	//ErrorCardNotReloadable  fUserErrorMessage = "Card not reloadable: %s"
	//ErrorRequiredParameter  fUserErrorMessage = "Missing required parameter %s"
	ErrorInvalidHeader      fUserErrorMessage = "Invalid header %s"
	//ErrorInvalidTopUpAmount fUserErrorMessage = "Invalid top-up amount for card %s"
	//ErrorMaxBalance         fUserErrorMessage = "Exceeded max balance for card %s"
	//ErrorNotCsvType         fUserErrorMessage = "File is not a CSV: %s"
	//ErrorMaxCountCards      fUserErrorMessage = "Maximum cards per order is %d"
)

type UserErrorMessage interface {
	UserMessage() string
}

type FormattedUserMessage interface {
	Format(args ...string) UserErrorMessage
}

type sUserErrorMessage string

func (s sUserErrorMessage) UserMessage() string {
	return string(s)
}

type fUserErrorMessage string

func (s fUserErrorMessage) Format(args ...interface{}) UserErrorMessage {
	return sUserErrorMessage(fmt.Sprintf(string(s), args...))
}

type Error interface {
	Status() int
	UserMessage() UserErrorMessage
	ValidationErrors() []string
	Error() string
}

func IsNotFoundError(err error) bool {
	if e, ok := err.(Error); ok {
		return e.Status() == http.StatusNotFound
	}
	return false
}

type httpError struct {
	status           int
	err              error
	userMessage      UserErrorMessage
	validationErrors []string
}

func (h *httpError) Error() string {
	return fmt.Sprintf("code: %d, userMessage: %s, validation: %v, err: %v", h.status, h.userMessage, h.validationErrors, h.err)
}

func (h *httpError) Status() int {
	return h.status
}

func (h *httpError) UserMessage() UserErrorMessage {
	return h.userMessage
}

func (h *httpError) ValidationErrors() []string {
	return h.validationErrors
}

type statusError struct {
	entity  string
	set     []string
	allowed bool
	status  string
}

func (i *statusError) Error() string {
	keyword := "disallowed"
	if i.allowed {
		keyword = "allowed"
	}
	return fmt.Sprintf("%s invalid status %s, %s = %v", i.entity, i.status, keyword, i.set)
}

func (i *statusError) Status() int {
	return http.StatusBadRequest
}

func (i *statusError) UserMessage() UserErrorMessage {
	return ErrorStatus.Format(i.entity, i.status)
}

func (i *statusError) ValidationErrors() []string {
	return nil
}

type contextualError struct {
	context string
	err     error
}

func (c *contextualError) Error() string {
	return fmt.Sprintf("%s: %v", c.context, c.err)
}

func (c *contextualError) Status() int {
	if e, ok := c.err.(Error); ok {
		return e.Status()
	}
	return http.StatusInternalServerError
}

func (c *contextualError) UserMessage() UserErrorMessage {
	if e, ok := c.err.(Error); ok {
		return e.UserMessage()
	}
	return ErrorInternal
}

func (c *contextualError) ValidationErrors() []string {
	if e, ok := c.err.(Error); ok {
		return e.ValidationErrors()
	}
	return nil
}

// Wraps another error to build up a stack of context
func ContextualError(err error, context string, formatArgs ...interface{}) error {
	c := fmt.Sprintf(context, formatArgs...)
	e := err
	if ce, ok := err.(*contextualError); ok {
		c = fmt.Sprintf("%s: %s", c, ce.context)
		e = ce.err
	}
	return &contextualError{
		context: c,
		err:     e,
	}
}

func BadError(message UserErrorMessage, err error) error {
	return NewHttpError(http.StatusBadRequest, message, err)
}

func NewHttpError(status int, message UserErrorMessage, err error) error {
	return &httpError{status: status, err: err, userMessage: message}
}

func StatusError(entity string, status string, required string) error {
	return &statusError{entity: entity, status: status, set: []string{required}, allowed: true}
}

func StatusErrorSet(entity string, status string, set []string, allowed bool) error {
	return &statusError{entity: entity, status: status, set: set, allowed: allowed}
}

func UnauthorizedError(err error) error {
	return NewHttpError(http.StatusUnauthorized, ErrorUnauthorized, err)
}

func ForbiddenError(err error) error {
	return NewHttpError(http.StatusForbidden, ErrorForbidden, err)
}

func NotImplementedError(err error) error {
	return NewHttpError(http.StatusNotImplemented, ErrorNotImplemented, err)
}

func ValidationErrors(err error, userErrors []string) error {
	return &httpError{
		status:           http.StatusBadRequest,
		err:              err,
		userMessage:      ErrorValidation,
		validationErrors: userErrors,
	}
}
