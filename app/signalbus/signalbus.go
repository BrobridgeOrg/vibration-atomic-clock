package signalbus

import (
	"time"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type SignalBus struct {
	host              string
	clientName        string
	client            *nats.Conn
	reconnectHandler  func(natsConn *nats.Conn)
	disconnectHandler func(natsConn *nats.Conn)
}

func CreateConnector(host string, clientName string, reconnectHandler func(natsConn *nats.Conn), disconnectHandler func(natsConn *nats.Conn)) *SignalBus {
	return &SignalBus{
		host:              host,
		clientName:        clientName,
		reconnectHandler:  reconnectHandler,
		disconnectHandler: disconnectHandler,
	}
}

func (sb *SignalBus) Connect() error {

	log.WithFields(log.Fields{
		"host":       sb.host,
		"clientName": sb.clientName,
	}).Info("Connecting to signal server")

	// Connect to signal server
	nc, err := nats.Connect(sb.host,
		nats.Name(sb.clientName),
		nats.PingInterval(10*time.Second),
		nats.MaxPingsOutstanding(3),
		nats.MaxReconnects(-1),
		nats.ReconnectHandler(sb.reconnectHandler),
		nats.DisconnectHandler(sb.disconnectHandler),
	)
	if err != nil {
		return err
	}

	sb.client = nc

	return nil
}

func (sb *SignalBus) Close() {
	sb.client.Close()
}

func (sb *SignalBus) Emit(topic string, data []byte) error {

	if err := sb.client.Publish(topic, data); err != nil {
		return err
	}

	return nil
}

func (sb *SignalBus) Watch(topic string, fn func(*nats.Msg)) (*nats.Subscription, error) {

	// Subscribe
	sub, err := sb.client.Subscribe(topic, fn)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (sb *SignalBus) QueueSubscribe(channelName string, topic string, fn func(*nats.Msg)) (*nats.Subscription, error) {

	// Subscribe
	sub, err := sb.client.QueueSubscribe(channelName, topic, fn)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (sb *SignalBus) Subscribe(topic string, fn func(*nats.Msg)) (*nats.Subscription, error) {

	// Subscribe
	sub, err := sb.client.Subscribe(topic, fn)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (sb *SignalBus) Unsubscribe(sub *nats.Subscription) error {

	// Unsubscribe
	err := sub.Unsubscribe()
	if err != nil {
		return err
	}

	return nil
}
