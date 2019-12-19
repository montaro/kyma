package testsuite

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// SendEvent is a step which sends example event to the application gateway
type SendEventToMesh struct {
	state   SendEventState
	appName string
}

var _ step.Step = &SendEventToMesh{}

// NewSendEvent returns new SendEvent
func NewSendEventToMesh(appName string, state SendEventState) *SendEventToMesh {
	return &SendEventToMesh{state: state, appName: appName}
}

// Name returns name name of the step
func (s *SendEventToMesh) Name() string {
	return "Send event"
}

// Run executes the step
func (s *SendEventToMesh) Run() error {
	event := s.prepareEvent()
	return s.state.GetEventSender().SendEventToMesh(s.appName, event)
}

func (s *SendEventToMesh) prepareEvent() *cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	data := "some data"
	event.SetID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	event.SetType(example_schema.EventType)
	event.SetSource("some source")
	event.SetData(data)
	event.SetTime(time.Now())
	event.SetExtension("eventtypeversion", example_schema.EventVersion)
	event.Validate()

	return &event
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToMesh) Cleanup() error {
	return nil
}
