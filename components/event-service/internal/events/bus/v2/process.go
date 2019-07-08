package v2

import (
	"fmt"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

type configurationData struct {
	SourceID string
}

//Conf Event-Service configuration data
var Conf *configurationData

//EventsTargetURL Event-Service contains target URL
var EventsTargetURL string

// Init should be used to initialize the "source" related configuration data
func Init(sourceID string, targetURL string) {
	Conf = &configurationData{
		SourceID: sourceID,
	}

	EventsTargetURL = targetURL
}

//CheckConf assert the configuration initialization
func CheckConf() (err error) {
	if Conf == nil {
		return fmt.Errorf("configuration data not initialized")
	}
	return nil
}

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParametersV3) (resp *api.SendEventParametersV3, err error) {
	if err := CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParametersV3{
		SourceID:            Conf.SourceID, // enrich the event with the sourceID
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
