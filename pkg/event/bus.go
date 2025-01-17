package event

import (
	"context"
	"errors"
	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"strings"
	"sync"
)

type GlobalEventBusClients struct {
	Mutex   sync.RWMutex
	Clients map[string]*ProducerConsumerClient
}

type ProducerConsumerClient struct {
	Producer *cloudevents.Client
	Consumer *cloudevents.Client
}

func (globalClient *GlobalEventBusClients) GetClient(endpoint string, subject string) (*ProducerConsumerClient, error) {
	// get from cache
	key := endpoint + subject
	globalClient.Mutex.RLock()
	clientP, ok := globalClient.Clients[key]
	globalClient.Mutex.RUnlock()
	if !ok {
		// build producer
		producerP, err := cenats.NewSender(endpoint, subject, cenats.NatsOptions())
		if err != nil {
			return nil, err
		}
		producerClient, err := cloudevents.NewClient(producerP)
		if err != nil {
			return nil, err
		}
		// build consumer
		consumerP, err := cenats.NewConsumer(endpoint, subject, cenats.NatsOptions())
		if err != nil {
			return nil, err
		}
		consumerClient, err := cloudevents.NewClient(consumerP)
		if err != nil {
			return nil, err
		}
		// update cache
		globalClient.Mutex.Lock()
		defer globalClient.Mutex.Unlock()
		if globalClient.Clients == nil {
			globalClient.Clients = make(map[string]*ProducerConsumerClient)
		}
		globalClient.Clients[key] = &ProducerConsumerClient{
			Producer: &producerClient,
			Consumer: &consumerClient,
		}
		clientP = globalClient.Clients[key]
	}
	return clientP, nil
}

var CurrentEventBusClient = &GlobalEventBusClients{}

type EventBus struct {
	Server  string
	Subject string
}

func (bus *EventBus) WithEndpoint(endpoint string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Server = endpoint
	return bus
}

func (bus *EventBus) WithSubject(subject string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Subject = strings.ToLower(subject)
	return bus
}

func (bus *EventBus) Publish(ctx context.Context, data interface{}) error {
	event, err := builderEvent(data)
	if err != nil {
		return err
	}
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	result := (*client.Producer).Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Subscribe(ctx context.Context, fn interface{}) error {
	client, err := CurrentEventBusClient.GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	return (*client.Consumer).StartReceiver(ctx, fn)
}
