package handlers

import (
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
)

// compile time check
var _ NatsInterface = &Nats{}

type NatsInterface interface {
	Initialize(cfg env.NatsConfig)
	DeleteSubscription(subscription *nats.Subscription)
}

type Nats struct {
	Client *nats.Conn
	Log    logr.Logger
}

type NatsResponse struct {
	StatusCode int
	Error      error
}

func (n *Nats) Initialize(cfg env.NatsConfig) {
	n.Log.Info("Initialize NATS connection")
	if n.Client == nil {
		var err error
		n.Client, err = nats.Connect(cfg.Url)
		if err != nil {
			n.Log.Error(err, "Can't connect to NATS Server")
		}
	}
	//TODO remove me, only for testing
	n.Client.Subscribe("foo", func(m *nats.Msg) {
		n.Log.Info("Received a message:", "message:", string(m.Data))
	})
}

func (n *Nats) DeleteSubscription(subscription *nats.Subscription) {
	n.Log.Info("Deleting NATS subscription...")
	subscription.Unsubscribe()
	panic("implement me")
}
