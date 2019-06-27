package v2

import (
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"regexp"
	"time"
)

var (
	isValidEventID = regexp.MustCompile(api.AllowedEventIDChars).MatchString

	// channel name components
	isValidSourceID         = regexp.MustCompile(api.AllowedSourceIDChars).MatchString
	isValidEventType        = regexp.MustCompile(api.AllowedEventTypeChars).MatchString
	isValidEventTypeVersion = regexp.MustCompile(api.AllowedEventTypeVersionChars).MatchString
)

//ValidatePublish validates a publish POST request
func ValidatePublish(r *EventRequestV3, opts *api.EventOptions) *api.Error {
	if len(r.Source) == 0 {
		return api.ErrorResponseMissingFieldSourceID()
	}
	if len(r.SpecVersion) == 0 {
		return ErrorResponseMissingFieldSpecVersion()
	}
	if len(r.Type) == 0 {
		return api.ErrorResponseMissingFieldEventType()
	}
	if len(r.TypeVersion) == 0 {
		return api.ErrorResponseMissingFieldEventTypeVersion()
	}
	if len(r.Time) == 0 {
		return api.ErrorResponseMissingFieldEventTime()
	}
	if r.Data == nil {
		return api.ErrorResponseMissingFieldData()
	} else if d, ok := (r.Data).(string); ok && d == "" {
		return api.ErrorResponseMissingFieldData()
	}

	//validate the event components lengths
	if len(r.Source) > opts.MaxSourceIDLength {
		return api.ErrorInvalidSourceIDLength(opts.MaxSourceIDLength)
	}
	if len(r.Type) > opts.MaxEventTypeLength {
		return api.ErrorInvalidEventTypeLength(opts.MaxEventTypeLength)
	}
	if len(r.TypeVersion) > opts.MaxEventTypeVersionLength {
		return api.ErrorInvalidEventTypeVersionLength(opts.MaxEventTypeVersionLength)
	}

	// validate the fully-qualified topic name components
	if !isValidSourceID(r.Source) {
		return api.ErrorResponseWrongSourceID(r.SourceIDFromHeader)
	}
	if !isValidEventType(r.Type) {
		return api.ErrorResponseWrongEventType()
	}
	if !isValidEventTypeVersion(r.TypeVersion) {
		return api.ErrorResponseWrongEventTypeVersion()
	}
	if r.SpecVersion!=SpecVersionV3 {
		return ErrorResponseWrongSpecVersion()
	}

	if _, err := time.Parse(time.RFC3339, r.Time); err != nil {
		return ErrorResponseWrongEventTime()
	}
	if len(r.ID) > 0 && !isValidEventID(r.ID) {
		return api.ErrorResponseWrongEventID()
	}
	return nil
}
