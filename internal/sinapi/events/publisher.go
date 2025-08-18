package events

import (
	"encoding/json"
	"go-confess-sins-api/pkg/models"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishSinUpdate(sin models.Sin) error {
	sinData, err := json.Marshal(sin)
	if err != nil {
		return err
	}
	return p.nc.Publish("sins.updated", sinData)
}
