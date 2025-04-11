package webhook

import "net/http"

type EventError interface {
	error
	SendHttpError(response http.ResponseWriter)
}

type InvalidEvent struct {
	Msg string
}

func (e *InvalidEvent) Error() string {
	return e.Msg
}

func (e *InvalidEvent) SendHttpError(response http.ResponseWriter) {
	http.Error(response, e.Error(), http.StatusBadRequest)
}

type FatalError struct {
	Msg string
}

func (e *FatalError) Error() string {
	return e.Msg
}

func (e *FatalError) SendHttpError(response http.ResponseWriter) {
	http.Error(response, e.Error(), http.StatusInternalServerError)
}

type MissingResourceError struct {
	Msg string
}

func (e *MissingResourceError) Error() string {
	return e.Msg
}

func (e *MissingResourceError) SendHttpError(response http.ResponseWriter) {
	http.Error(response, e.Error(), http.StatusGone)
}
