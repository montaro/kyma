package v2

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

// ErrorResponseMissingFieldEventID returns an error of type PublishEventResponses for missing EventID field
func ErrorResponseMissingFieldEventID() (response *api.PublishEventResponses) {
	return shared.CreateMissingFieldError(shared.FieldEventIDV3)
}

// ErrorResponseMissingFieldEventType returns an error of type PublishEventResponses for missing EventType field
func ErrorResponseMissingFieldEventType() (response *api.PublishEventResponses) {
	return shared.CreateMissingFieldError(shared.FieldEventTypeV3)
}

// ErrorResponseMissingFieldEventTypeVersion returns an error of type PublishEventResponses for missing EventTypeVersion field
func ErrorResponseMissingFieldEventTypeVersion() (response *api.PublishEventResponses) {
	return shared.CreateMissingFieldError(shared.FieldEventTypeVersionV3)
}

// ErrorResponseWrongEventTypeVersion returns an error of type PublishEventResponses for wrong EventTypeVersion field
func ErrorResponseWrongEventTypeVersion() (response *api.PublishEventResponses) {
	return shared.CreateInvalidFieldError(shared.FieldEventTypeVersionV3)
}

// ErrorResponseMissingFieldEventTime returns an error of type PublishEventResponses for missing EventTime field
func ErrorResponseMissingFieldEventTime() (response *api.PublishEventResponses) {
	return shared.CreateMissingFieldError(shared.FieldEventTimeV3)
}

// ErrorResponseWrongEventTime returns an error of type PublishEventResponses for wrong EventTime field
func ErrorResponseWrongEventTime() (response *api.PublishEventResponses) {
	return shared.CreateInvalidFieldError(shared.FieldEventTimeV3)
}

// ErrorResponseWrongEventID returns an error of type PublishEventResponses for wrong EventID field
func ErrorResponseWrongEventID() (response *api.PublishEventResponses) {
	return shared.CreateInvalidFieldError(shared.FieldEventIDV3)
}

// ErrorResponseWrongSpecVersion returns an error of type PublishEventResponses for wrong SpecVersion field
func ErrorResponseWrongSpecVersion() (response *api.PublishEventResponses) {
	return shared.CreateInvalidFieldError(shared.FieldSpecVersionV3)
}

// ErrorResponseMissingFieldData returns an error of type PublishEventResponses for missing Data field
func ErrorResponseMissingFieldData() (response *api.PublishEventResponses) {
	return shared.CreateMissingFieldError(shared.FieldData)
}
