package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// type EventLog type {
// 	ProductID int `json:"product_id"`
// 	PreviousQuantity int `json:"previous_quantity"`
// 	NewQuantity int `json:"new_quantity`
// 	EventType string `json:event_type`
// }

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"inventory",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare exchange")

	queue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare queue")

	err = ch.QueueBind(
		queue.Name,
		"inventory.updates",
		"inventory",
		false,
		nil,
	)
	failOnError(err, "Failed to bind queue")

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	fmt.Println("Connection started, waiting to receive (CTRL+C to exit)")
	for msg := range msgs {
		// eventType := EvaluateThresholds(msg)
	}

	// loop declaration
	<-make(chan struct{})
}

// func EvaluateThresholds(msg []byte) string {
//
// }
