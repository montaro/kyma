package bus

import (
	"fmt"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

type configurationData struct {
	SourceID string
}

//Conf Event-Service configuration data
var Conf *configurationData
var eventsTargetURL string

// Init should be used to initialize the "source" related configuration data
func Init(sourceID string, targetURL string) {
	Conf = &configurationData{
		SourceID: sourceID,
	}

	eventsTargetURL = targetURL
}

//CheckConf assert the configuration initialization
func CheckConf() (err error) {
	if Conf == nil {
		return fmt.Errorf("configuration data not initialized")
	}
	return nil
}

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParameters) (resp *api.SendEventParameters, err error) {
	if err := CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParameters{
		SourceID:         Conf.SourceID, // enrich the event with the sourceID
		EventType:        parameters.Publishrequest.EventType,
		EventTypeVersion: parameters.Publishrequest.EventTypeVersion,
		EventID:          parameters.Publishrequest.EventID,
		EventTime:        parameters.Publishrequest.EventTime,
		Data:             parameters.Publishrequest.Data,
	}

	return &sendRequest, nil
}
