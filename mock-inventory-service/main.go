package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	//TODO: update this to take from os.Getenv("RABBITMQ_URL")
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
	failOnError(err, "Failed to declare an exchange")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	body, err := bodyFrom(os.Args)
	failOnError(err, "Failed to create message body")

	err = ch.PublishWithContext(ctx,
		"inventory",
		"inventory.updates",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf("Sent: %s", string(body))
}

type NoArgsError struct{}
type OneArgError struct{}

func (b *NoArgsError) Error() string {
	return "No arguments provided (expected args 'productID quantity')"
}
func (b *OneArgError) Error() string {
	return "Only one argument provided (expected args 'productID quantity')"
}

func bodyFrom(args []string) (body []byte, err error) {
	if len(args) < 2 {
		return nil, &NoArgsError{}
	} else if len(args) == 2 {
		return nil, &OneArgError{}
	}

	productID := os.Args[1]
	quantity := os.Args[2]
	timestamp := time.Now()

	jsonString := fmt.Sprintf(`{"product_id":"%s","quantity":%s,"timestamp":"%s"}`, productID, quantity, timestamp)
	jsonBody, err := json.Marshal(jsonString)

	return jsonBody, nil
}
