package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/pkg/errors"

	cloudevents "github.com/cloudevents/sdk-go"

)

type EventSender struct {
	httpClient resilient.HttpClient
	domain     string
}

func NewEventSender(httpClient resilient.HttpClient, domain string) *EventSender {
	return &EventSender{
		httpClient: httpClient,
		domain:     domain,
	}
}

func (s *EventSender) SendEvent(appName string, event *ExampleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gateway.%s/%s/v1/events", s.domain, appName)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("send event failed: %v\nrequest: %v\nresponse: %v", response.StatusCode, request, response)
	}

	return nil
}

func (s *EventSender) SendEventToMesh(appName string, event *ExampleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gateway.%s/%s/events", s.domain, appName)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(url),
	)
	t.Client = s.httpClient

	if err != nil {
		panic("failed to create transport, " + err.Error())
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		panic("unable to create cloudevent client: " + err.Error())
	}
	if err := c.Send(ctx, event); err != nil {
		panic("failed to send cloudevent: " + err.Error())
	}


	url := fmt.Sprintf("https://gateway.%s/%s/events", s.domain, appName)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("send event to the Knative Event Mesh failed: %v\nrequest: %v\nresponse: %v", response.StatusCode, request, response)
	}

	return nil
}
