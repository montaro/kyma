package v2

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
)

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParametersV3) (resp *api.SendEventParametersV3, err error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParametersV3{
		SourceID:            bus.Conf.SourceID, // enrich the event with the sourceID
		EventType:           parameters.EventRequestV3.EventType,
		EventTypeVersion:    parameters.EventRequestV3.EventTypeVersion,
		EventID:             parameters.EventRequestV3.EventID,
		EventTime:           parameters.EventRequestV3.EventTime,
		SpecVersion:         parameters.EventRequestV3.SpecVersion,
		DataContentEncoding: parameters.EventRequestV3.DataContentEncoding,
		Data:                parameters.EventRequestV3.Data,
	}

	return &sendRequest, nil
}
