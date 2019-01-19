package mailconsumer

import (
	"bytes"
	"fmt"
	"html/template"
)

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shal/pigeon/pkg/consumer"
	"github.com/shal/pigeon/pkg/eventapi"
	"github.com/shal/pigeon/pkg/utils"
	"github.com/streadway/amqp"
)

const (
	routingKey = "user.email.confirmation.token"
	exchange   = "barong.events.system"
)

func amqpURI() string {
	host := utils.GetEnv("RABBITMQ_HOST", "localhost")
	port := utils.GetEnv("RABBITMQ_PORT", "5672")
	username := utils.GetEnv("RABBITMQ_USERNAME", "guest")
	password := utils.GetEnv("RABBITMQ_PASSWORD", "guest")

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)
}

type Consumer struct{}

func (Consumer) Handle(event eventapi.Event) error {
	// Decode map[string]interface{} to AccountRecord.
	acc := AccountCreatedEvent{}

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           &acc,
		WeaklyTypedInput: true,
	})

	if err != nil {
		return err
	}

	if err := dec.Decode(event); err != nil {
		return err
	}

	tpl, err := template.ParseFiles("templates/sign_up.tpl")
	if err != nil {
		return err
	}

	buff := bytes.Buffer{}
	if err := tpl.Execute(&buff, acc); err != nil {
		return err
	}

	apiKey := utils.MustGetEnv("SENDGRID_API_KEY")

	email := eventapi.Email{
		FromAddress: utils.GetEnv("SENDER_EMAIL", "noreply@pigeon.com"),
		FromName:    utils.GetEnv("SENDER_NAME", "Pigeon"),
		Subject:     "Confirmation Instructions",
		Reader:      bytes.NewReader(buff.Bytes()),
	}

	if _, err := email.Send(apiKey, acc.User.Email); err != nil {
		return err
	}

	return nil
}

func (Consumer) Consume() (<-chan amqp.Delivery, error) {
	amqpUri := amqpURI()

	c := consumer.New(amqpUri, exchange, routingKey)
	queue := c.DeclareQueue()
	c.BindQueue(queue)

	return c.Channel.Consume(
		queue.Name,
		c.Tag,
		true,
		true,
		false,
		false,
		nil,
	)
}
