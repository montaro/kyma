package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/stretchr/testify/assert"
)

const (
	// TestSource used in the tests
	TestSource = "test-source"
	// TestEventType used in the tests
	TestEventID = "4ea567cf-812b-49d9-a4b2-cb5ddf464094"
	// TestType used in the tests
	TestType = "test-type"
	// TestEventTypeVersion used in the tests
	TestEventTypeVersion = "v1"
	// TestEventTypeVersionInvalid used in the tests
	TestEventTypeVersionInvalid = "#"
	// TestEventTime used in the tests
	TestEventTime = "2012-11-01T22:08:41+00:00"
	// TestEventTimeInvalid used in the tests
	TestEventTimeInvalid = "2012-11-01T22"
	// TestEventIDInvalid used in the tests
	TestEventIDInvalid = "4ea567cf"
	// TestSourceIDInvalid used in the tests
	TestSourceIDInvalid = "source/Id"
	// TestSpecVersion used in the tests
	TestSpecVersion = "0.3"

	// TestSpecVersionInvalid used in the tests
	TestSpecVersionInvalid = "0.2"

	// TestData used in the tests
	TestData = "{'key':'value'}"
	// TestDataEmpty used in the tests
	TestDataEmpty = ""

	// event payload format
	eventFormat            = "{%v}"
	eventIDFormat          = "\"id\":\"%v\""
	eventTypeFormat        = "\"type\":\"%v\""
	eventTypeVersionFormat = "\"eventtypeversion\":\"%v\""
	eventTimeFormat        = "\"time\":\"%v\""
	eventSpecVersion       = "\"specversion\":\"%v\""
	source                 = "\"source\":\"%v\""
	contentType            = "\"content-type:\":\"%v\""
	dataFormat             = "\"data\":\"%v\""
)

type eventBuilder struct {
	bytes.Buffer
}

func (b *eventBuilder) build(format string, values ...interface{}) *eventBuilder {
	if b.Len() > 0 {
		_, _ = b.WriteString(",")
	}
	source := fmt.Sprintf(format, values...)
	_, _ = b.WriteString(source)
	return b
}

// BuildV2PayloadWithoutCEFields returns a payload without any Cloud Event fields
func BuildV2PayloadWithoutCEFields() string {
	builder := new(eventBuilder).
		build("", "")
	payload := builder.String()
	return payload
}

func (b *eventBuilder) String() string {
	return fmt.Sprintf(eventFormat, b.Buffer.String())
}

// BuildPublishV2TestPayload returns a complete payload compliant with CE 0.3
func BuildPublishV2TestPayload(sourceID, eventType, eventTypeVersion, eventID, eventTime, data string) string {
	builder := new(eventBuilder).
		build(source, sourceID).
		build(eventSpecVersion, TestSpecVersion).
		build(eventTypeFormat, eventType).
		build(eventTypeVersionFormat, eventTypeVersion).
		build(eventIDFormat, eventID).
		build(eventTimeFormat, eventTime).
		build(dataFormat, data)
	payload := builder.String()
	return payload
}

// BuildPublishV2TestPayloadWithInvalidSpecversion returns a payload with invalid specversion
func BuildPublishV2TestPayloadWithInvalidSpecversion() string {
	builder := new(eventBuilder).
		build(eventSpecVersion, TestSpecVersionInvalid).
		build(eventIDFormat, TestEventID).
		build(source, TestSource).
		build(eventTypeFormat, TestType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildPublishV2TestPayloadWithoutID returns a complete payload compliant with CE 0.3
func BuildPublishV2TestPayloadWithoutID() string {
	builder := new(eventBuilder).
		build(eventSpecVersion, TestSpecVersion).
		build(source, TestSource).
		build(eventTypeFormat, TestType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildPublishV2TestPayloadWithoutSource returns a complete payload compliant with CE 0.3
func BuildPublishV2TestPayloadWithoutSource() string {
	builder := new(eventBuilder).
		build(eventSpecVersion, TestSpecVersion).
		build(eventIDFormat, TestEventID).
		build(eventTypeFormat, TestType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildPublishV2TestPayloadWithoutSpecVersion returns a complete payload compliant with CE 0.3
func BuildPublishV2TestPayloadWithoutSpecversion() string {
	builder := new(eventBuilder).
		build(eventIDFormat, TestEventID).
		build(eventTypeFormat, TestType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildPublishV2TestPayloadWithoutType returns a complete payload compliant with CE 0.3
func BuildPublishV2TestPayloadWithoutType() string {
	builder := new(eventBuilder).
		build(eventSpecVersion, TestSpecVersion).
		build(eventIDFormat, TestEventID).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// PerformPublishV2RequestWithHeaders performs a test publish request with HTTP headers.
func PerformPublishV2RequestWithHeaders(t *testing.T, publishURL string, payload string) ([]byte, int) {
	req, _ := http.NewRequest("POST", publishURL+"/v2/events", strings.NewReader(payload))

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	return body, res.StatusCode
}

// AssertExpectedError asserts an expected status-code and error.
func AssertExpectedError(t *testing.T, body []byte, actualStatusCode int, expectedStatusCode int, errorField interface{}, errorType interface{}) {
	var responseError api.Error
	err := json.Unmarshal(body, &responseError)
	assert.Equal(t, expectedStatusCode, actualStatusCode)
	assert.Nil(t, err)
	if errorType != nil {
		assert.Equal(t, errorType, responseError.Type)
	}
	if errorField != nil {
		assert.NotNil(t, responseError.Details)
		assert.NotEqual(t, len(responseError.Details), 0)
		assert.Equal(t, errorField, responseError.Details[0].Field)
	}
}

// // BuildDefaultTestPayload returns a default test payload.
// func BuildDefaultTestPayload() string {
// 	return BuildTestPayload(TestSourceID, TestEventType, TestEventTypeVersion, TestEventID, TestEventTime, TestData)
// }

// // BuildDefaultTestBadPayload returns a default test bad payload.
// func BuildDefaultTestBadPayload() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeFormat, TestEventType).
// 		build(eventTypeVersionFormat, TestEventTypeVersion).
// 		build(eventIDFormat, TestEventID).
// 		build(eventTimeFormat, TestEventTime).
// 		build(dataFormat, TestData)
// 	_, _ = builder.WriteString(",") // spoil the payload
// 	payload := builder.String()
// 	return payload
// }

// // BuildDefaultTestPayloadWithoutEventType returns a default test payload without the event-type.
// func BuildDefaultTestPayloadWithoutEventType() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeVersionFormat, TestEventTypeVersion).
// 		build(eventIDFormat, TestEventID).
// 		build(eventTimeFormat, TestEventTime).
// 		build(dataFormat, TestData)
// 	payload := builder.String()
// 	return payload
// }

// // BuildDefaultTestPayloadWithoutEventTypeVersion returns a default test payload without the event-type-version.
// func BuildDefaultTestPayloadWithoutEventTypeVersion() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeFormat, TestEventType).
// 		build(eventIDFormat, TestEventID).
// 		build(eventTimeFormat, TestEventTime).
// 		build(dataFormat, TestData)
// 	payload := builder.String()
// 	return payload
// }

// // BuildDefaultTestPayloadWithoutEventTime returns a default test payload without the event-time
// func BuildDefaultTestPayloadWithoutEventTime() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeFormat, TestEventType).
// 		build(eventTypeVersionFormat, TestEventTypeVersion).
// 		build(eventIDFormat, TestEventID).
// 		build(dataFormat, TestData)
// 	payload := builder.String()
// 	return payload
// }

// // BuildDefaultTestPayloadWithoutData returns a default test payload without the data
// func BuildDefaultTestPayloadWithoutData() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeFormat, TestEventType).
// 		build(eventTypeVersionFormat, TestEventTypeVersion).
// 		build(eventIDFormat, TestEventID).
// 		build(eventTimeFormat, TestEventTime)
// 	payload := builder.String()
// 	return payload
// }

// // BuildDefaultTestPayloadWithEmptyData returns a default test payload with empty data
// func BuildDefaultTestPayloadWithEmptyData() string {
// 	builder := new(eventBuilder).
// 		build(sourceIDFormat, TestSourceID).
// 		build(eventTypeFormat, TestEventType).
// 		build(eventTypeVersionFormat, TestEventTypeVersion).
// 		build(eventIDFormat, TestEventID).
// 		build(eventTimeFormat, TestEventTime).
// 		build(dataFormat, TestDataEmpty)
// 	payload := builder.String()
// 	return payload
// }

// // PerformPublishV1Request performs a test publish request.
// func PerformPublishV1Request(t *testing.T, publishURL string, payload string) ([]byte, int) {
// 	res, err := http.Post(publishURL+"/v1/events", "application/json", strings.NewReader(payload))

// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer func() { _ = res.Body.Close() }()
// 	body, err := ioutil.ReadAll(res.Body)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	return body, res.StatusCode
// }
