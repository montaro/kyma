package v2

import (
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
)
// ErrorResponseMissingFieldEventType returns an API error instance for the missing field event type error.
func ErrorResponseMissingFieldSpecVersion() (response *api.Error) {
	return api.CreateMissingFieldError(FieldSpecVersion)
}

// ErrorResponseWrongEventType returns an API error instance for the wrong event type error.
func ErrorResponseWrongSpecVersion() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldSpecVersion)
}

// ErrorResponseWrongEventTime returns an API error instance for the wrong event time error.
func ErrorResponseWrongEventTime() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldEventTime)
}