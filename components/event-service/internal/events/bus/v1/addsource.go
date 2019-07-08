package v1

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
)

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParameters) (resp *api.SendEventParameters, err error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParameters{
		SourceID:         bus.Conf.SourceID, // enrich the event with the sourceID
		EventType:        parameters.Publishrequest.EventType,
		EventTypeVersion: parameters.Publishrequest.EventTypeVersion,
		EventID:          parameters.Publishrequest.EventID,
		EventTime:        parameters.Publishrequest.EventTime,
		Data:             parameters.Publishrequest.Data,
	}

	return &sendRequest, nil
}
