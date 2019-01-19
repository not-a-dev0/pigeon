package main

import (
	"github.com/shal/pigeon/pkg/consumer"
	"github.com/shal/pigeon/pkg/mailconsumer"
)

func main() {
	mailConsumer := mailconsumer.Consumer{}
	consumer.ListenAndServe(mailConsumer)
}
