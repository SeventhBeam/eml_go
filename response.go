package eml

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Response interface {
	http.ResponseWriter
	Text(statusCode int, body string) Response
	Json(statusCode int, model interface{}) Response
	JsonOk(model interface{}) Response
	Error(statusCode int, message UserErrorMessage, validations []string) Response
	CacheSeconds(seconds int) Response
	CacheMinutes(minutes int) Response
	HandleJsonOk(model interface{}, err error) error
	HandleError(err interface{})
}

type emlResponse struct {
	http.ResponseWriter
}

func NewResponse(w http.ResponseWriter) Response {
	return &emlResponse{w}
}

type ErrorResponse struct {
	Message          string   `json:"message"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

func (r *emlResponse) Text(statusCode int, body string) Response {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(statusCode)
	fmt.Fprintf(r.ResponseWriter, "%s\n", body)
	return r
}

func (r *emlResponse) Json(statusCode int, model interface{}) Response {
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(statusCode)
	json.NewEncoder(r.ResponseWriter).Encode(model)
	return r
}

func (r *emlResponse) JsonOk(model interface{}) Response {
	return r.Json(http.StatusOK, model)
}

func (r *emlResponse) Error(statusCode int, message UserErrorMessage, validations []string) Response {
	return r.Json(statusCode, ErrorResponse{
		Message:          message.UserMessage(),
		ValidationErrors: validations,
	})
}

func (r *emlResponse) CacheSeconds(seconds int) Response {
	r.ResponseWriter.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", seconds))
	r.ResponseWriter.Header().Set("Vary", fmt.Sprintf("Authorization,x-api-key"))
	return r
}

func (r *emlResponse) CacheMinutes(minutes int) Response {
	return r.CacheSeconds(minutes * 60)
}

func (r *emlResponse) HandleJsonOk(model interface{}, err error) error {
	if err != nil {
		return err
	}
	r.JsonOk(model)
	return nil
}

func (r *emlResponse) HandleError(err interface{}) {
	log.Println(err)
	if e, ok := err.(Error); ok {
		r.Error(e.Status(), e.UserMessage(), e.ValidationErrors())
	} else {
		r.Error(http.StatusInternalServerError, ErrorInternal, nil)
	}
}
