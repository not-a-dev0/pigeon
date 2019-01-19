package consumer

import (
	"io/ioutil"
	"log"
)

import (
	"github.com/shal/pigeon/pkg/eventapi"
	"github.com/shal/pigeon/pkg/utils"
	"github.com/streadway/amqp"
)

const (
	tag       = "pigeon"
	queueName = "pigeon.events.consumer"
)

type Consumer struct {
	Conn       *amqp.Connection
	Channel    *amqp.Channel
	Tag        string
	RoutingKey string
	Exchange   string
}

func (c *Consumer) BindQueue(queue amqp.Queue) {
	err := c.Channel.QueueBind(
		queue.Name,
		c.RoutingKey,
		c.Exchange,
		false,
		nil,
	)

	if err != nil {
		log.Panicf("Queue Bind: %s", err.Error())
	}
}

func (c *Consumer) DeclareQueue() amqp.Queue {
	err := c.Channel.ExchangeDeclare(
		c.Exchange,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Panicf("Exchange: %s", err.Error())
	}

	queue, err := c.Channel.QueueDeclare(
		queueName,
		true,
		true,
		true,
		false,
		nil,
	)

	if err != nil {
		log.Panicf("Queue: %s", err.Error())
	}

	return queue
}

func New(uri, exchange, key string) *Consumer {
	// Create a connection.
	conn, err := amqp.Dial(uri)

	if err != nil {
		log.Panicf("Dial %s", err.Error())
	} else {
		log.Printf("Successfully connected to %s\n", uri)
	}

	// Create a channel.
	channel, err := conn.Channel()

	if err != nil {
		log.Panicf("Channel %s", err.Error())
	}

	consumer := &Consumer{
		Conn:       conn,
		Channel:    channel,
		Tag:        tag,
		Exchange:   exchange,
		RoutingKey: key,
	}

	return consumer
}

func ListenAndServe(handlers ...eventapi.EventHandler) {
	utils.MustGetEnv("JWT_PUBLIC_KEY")
	utils.MustGetEnv("SENDGRID_API_KEY")

	for _, handler := range handlers {
		deliveries, err := handler.Consume()

		if err != nil {
			log.Panicf("Consuming: %s", err.Error())
		}

		go func() {
			for delivery := range deliveries {
				jwtReader, err := eventapi.DeliveryAsJWT(delivery)

				if err != nil {
					log.Println(err)
					return
				}

				jwt, err := ioutil.ReadAll(jwtReader)
				if err != nil {
					log.Println(err)
					return
				}

				log.Printf("Token: %s\n", string(jwt))

				claims, err := eventapi.ParseJWT(string(jwt), eventapi.ValidateJWT)
				if err != nil {
					log.Println(err)
					return
				}

				if err := handler.Handle(claims.Event); err != nil {
					log.Printf("Consuming: %s\n", err.Error())
				}
			}
		}()
	}

	forever := make(chan bool)
	log.Printf("[*] Waiting for events. To exit press CTRL+C")
	<-forever
}
