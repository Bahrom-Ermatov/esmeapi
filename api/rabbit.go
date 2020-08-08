package api

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type rabbit struct {
	c  *amqp.Connection
	ch *amqp.Channel
	q  amqp.Queue
}

// NewRabbit get new connection with queue and channel
func newRabbit() (*rabbit, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to connect to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open a channel")
	}

	q, err := ch.QueueDeclare(
		channel, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to declare a queue")
	}

	return &rabbit{c: conn, ch: ch, q: q}, nil
}

// Publish message to rabbitmq
func (s *Server) Publish(data []byte) error {

	err := s.r.ch.Publish(
		"",         // exchange
		s.r.q.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})

	if err != nil {
		return errors.Wrap(err, "cannot publish message")
	}

	return nil
}
