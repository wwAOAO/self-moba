package nats

import (
	"time"

	natsgo "github.com/nats-io/nats.go"
)

type Client struct {
	conn *natsgo.Conn
}

func Connect(url string, name string) (*Client, error) {
	conn, err := natsgo.Connect(
		url,
		natsgo.Name(name),
		natsgo.Timeout(5*time.Second),
		natsgo.ReconnectWait(time.Second),
		natsgo.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Conn() *natsgo.Conn {
	return c.conn
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Drain()
	c.conn.Close()
}
