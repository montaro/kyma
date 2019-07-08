package api

// PublishRequestV1 implements the service definition of PublishRequest
type PublishRequestV1 struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// EventRequestV2 implements the service definition of EventRequestV3
type EventRequestV2 struct {
	EventType           string   `json:"type"`
	EventTypeVersion    string   `json:"eventtypeversion"`
	EventID             string   `json:"id"`
	EventTime           string   `json:"time"`
	SpecVersion         string   `json:"specversion"`
	DataContentEncoding string   `json:"datacontentencoding,omitempty"`
	Data                AnyValue `json:"data"`
}

// PublishResponse implements the service definition of PublishResponse
type PublishResponse struct {
	EventID string `json:"event-id,omitempty"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface {
}

// Error implements the service definition of APIError
type Error struct {
	Status   int           `json:"status,omitempty"`
	Type     string        `json:"type,omitempty"`
	Message  string        `json:"message,omitempty"`
	MoreInfo string        `json:"moreInfo,omitempty"`
	Details  []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail implements the service definition of APIErrorDetail
type ErrorDetail struct {
	Field    string `json:"field,omitempty"`
	Type     string `json:"type,omitempty"`
	Message  string `json:"message,omitempty"`
	MoreInfo string `json:"moreInfo,omitempty"`
}

// PublishEventParametersV1 holds parameters to PublishEvent
type PublishEventParametersV1 struct {
	PublishrequestV1 PublishRequestV1 `json:"publishrequest,omitempty"`
}

// PublishEventParametersV2 holds parameters to PublishEvent
type PublishEventParametersV2 struct {
	EventRequestV2 EventRequestV2 `json:"publishrequest,omitempty"`
}

// PublishEventResponses holds responses of PublishEvent
type PublishEventResponses struct {
	Ok    *PublishResponse
	Error *Error
}

// SendEventParametersV1 implements the request to the outbound messaging API
type SendEventParametersV1 struct {
	SourceID         string   `json:"source-id,omitempty"`
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// SendEventParametersV2 implements the request to the outbound messaging API
type SendEventParametersV2 struct {
	Source              string   `json:"source"`
	Type                string   `json:"type"`
	EventTypeVersion    string   `json:"eventtypeversion"`
	ID                  string   `json:"id"`
	Time                string   `json:"time"`
	SpecVersion         string   `json:"specversion"`
	DataContentEncoding string   `json:"datacontentencoding"`
	Data                AnyValue `json:"data"`
}

// SendEventResponse holds the response from outbound messaging API
type SendEventResponse PublishEventResponses
