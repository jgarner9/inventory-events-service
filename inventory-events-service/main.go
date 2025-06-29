package main

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type InventoryEvent struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Timestamp string `json:"timestamp"`
}

type EventLog struct {
	ProductID        string `json:"product_id"`
	PreviousQuantity int    `json:"previous_quantity"`
	NewQuantity      int    `json:"new_quantity"`
	EventType        string `json:"event_type"`
}

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
	var lastKnownQuantities = make(map[string]int)
	lastKnownQuantities["1"] = 11
	lastKnownQuantities["2"] = 1
	lastKnownQuantities["3"] = 0
	for msg := range msgs {
		fmt.Println("Message received")
		inventoryEventBody := GetBody(msg.Body)
		failOnError(err, "Unable to unmarshal JSON")

		productID := inventoryEventBody.ProductID
		previousQuantity := lastKnownQuantities[productID]
		currentQuantity := inventoryEventBody.Quantity
		eventType := EvaluateThresholds(currentQuantity, previousQuantity)

		if eventType != "" {
			err = CreateLog(EventLog{
				ProductID:        productID,
				PreviousQuantity: previousQuantity,
				NewQuantity:      currentQuantity,
				EventType:        eventType,
			})
			failOnError(err, "Failed to create log")
			fmt.Println("Log created")
		}

	}

	// loop declaration
	<-make(chan struct{})
}

func EvaluateThresholds(currentQuantity, previousQuantity int) string {
	if currentQuantity <= 0 && previousQuantity > 0 {
		return "OUT_OF_STOCK"
	} else if currentQuantity < 10 && previousQuantity >= 10 {
		return "LOW_STOCK"
	} else if currentQuantity > 0 && previousQuantity <= 0 {
		return "BACK_IN_STOCK"
	}

	return ""
}

func GetBody(body []byte) InventoryEvent {
	var bodyString string
	err := json.Unmarshal(body, &bodyString)
	failOnError(err, "Failed to unmarshal body to string")

	var jsonBody InventoryEvent
	err = json.Unmarshal([]byte(bodyString), &jsonBody)
	failOnError(err, "Failed to unmarshal body to InventoryEvent type")

	return jsonBody
}

func CreateLog(log EventLog) error {
	return nil
}
