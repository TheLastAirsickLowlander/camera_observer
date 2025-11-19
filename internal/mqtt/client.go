package mqtt

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	inner paho.Client
}

func NewClient(broker, clientID string) (*Client, error) {
	opts := paho.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(c paho.Client, err error) {
		fmt.Printf("[MQTT] Connection lost: %v\n", err)
	})
	opts.SetOnConnectHandler(func(c paho.Client) {
		fmt.Println("[MQTT] Connected")
	})

	c := paho.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Client{inner: c}, nil
}

func (c *Client) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	token := c.inner.Subscribe(topic, 1, func(client paho.Client, msg paho.Message) {
		handler(msg.Topic(), msg.Payload())
	})

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("[MQTT] Subscribed to %s\n", topic)
	return nil
}

func (c *Client) Disconnect() {
	c.inner.Disconnect(250)
}
