package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"time"
)

// SendEvent is a step which sends example event to the application gateway
type SendEventToMesh struct {
	state SendEventState
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

func (s *SendEventToMesh) prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        example_schema.EventType,
		EventTypeVersion: example_schema.EventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             "some data",
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToMesh) Cleanup() error {
	return nil
}
