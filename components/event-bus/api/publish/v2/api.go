package v2

const (
	// FieldData value
	FieldData = "data"
	// FieldEventID value
	FieldEventID = "id"
	// FieldEventTime value
	FieldEventTime = "time"
	// FieldEventType value
	FieldEventType = "type"
	// FieldSpecVersion value
	FieldSpecVersion = "specversion"
	// FieldEventTypeVersion value
	FieldEventTypeVersion = "eventtypeversion"
	// FieldSourceID value
	FieldSourceID = "source"
	//// FieldTraceContext value
	//FieldTraceContext = "trace-context"

	// AllowedEventIDChars regex
	AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`

	// AllowedSourceIDChars regex
	AllowedSourceIDChars = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	// AllowedEventTypeChars regex
	AllowedEventTypeChars = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	// AllowedEventTypeVersionChars regex
	AllowedEventTypeVersionChars = `^[a-zA-Z0-9]+$`

	// HeaderSourceID heaver
	HeaderSourceID = "Source"

	//SpecVersionV3.0 Value
	SpecVersionV3 = "0.3"
)
// Extensions type
type Extensions = map[string]interface{}

// AnyValue implements the service definition of AnyValue
type AnyValue interface{}

// EventRequestV3 represents a publish event CE v.3.0 request
type EventRequestV3 struct {
	ID                  string   `json:"id"`
	Source              string   `json:"source"`
	SpecVersion         string   `json:"specversion"`
	Type                string   `json:"type"`
	DataContentEncoding string   `json:"datacontentencoding,omitempty"`
	TypeVersion         string   `json:"eventtypeversion"`
	Time                string   `json:"time"`
	Data                AnyValue `json:"data"`
	SourceIDFromHeader  bool
}

// CloudEvent represents the event to be persisted to NATS
type CloudEventV3 struct {
	EventRequestV3
	Extensions Extensions `json:"extensions,omitempty"`
}
