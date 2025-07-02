package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

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

	var lastKnownQuantities = make(map[string]int)
	lastKnownQuantities["f47ac10b-58cc-4372-a567-0e02b2c3d479"] = 11
	lastKnownQuantities["2"] = 1
	lastKnownQuantities["3"] = 0

	// start internal API
	go func() {
		router := SetupAPIRouter()
		router.Run(":5665")
	}()

	for msg := range msgs {
		inventoryEventBody := GetBody(msg.Body)
		failOnError(err, "Unable to unmarshal JSON")

		productID := inventoryEventBody.ProductID
		previousQuantity := lastKnownQuantities[productID]
		currentQuantity := inventoryEventBody.Quantity
		eventType := EvaluateThresholds(currentQuantity, previousQuantity)

		if eventType != "" {
			CreateLog(EventLog{
				ProductID:        productID,
				PreviousQuantity: previousQuantity,
				NewQuantity:      currentQuantity,
				EventType:        eventType,
			})
		}

	}
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

func CreateLog(log EventLog) {
	ctx := context.Background()
	//TODO: Change this to take the connection string from os.Getenv()
	conn, err := pgx.Connect(ctx, "postgresql://invent:PurpleOctopi*22@localhost:5432/inventory_service?sslmode=disable")
	failOnError(err, "Failed to connect to DB")
	defer conn.Close(ctx)

	query := "INSERT INTO inventory_events (product_id, previous_quantity, new_quantity, event_type) VALUES ($1, $2, $3, $4);"

	status, err := conn.Exec(ctx, query, log.ProductID, log.PreviousQuantity, log.NewQuantity, log.EventType)
	failOnError(err, "Failed to add log to table")
	if status.RowsAffected() == 1 {
		fmt.Println("Event logged")
	} else {
		fmt.Printf("Something went wrong: %s\n", status.String())
	}
}

func SetupAPIRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/events", func(ctx *gin.Context) {
		productID := ctx.Query("product_id")
		limit := ctx.DefaultQuery("limit", "20")

		logs := GetLogs(productID, limit)

		if logs != "" {
			ctx.JSON(http.StatusOK, logs)
		} else {
			ctx.String(http.StatusNotFound, "No logs found for product ID")
		}
	})

	return router
}

func GetLogs(id, limit string) string {
	ctx := context.Background()
	//TODO: Change this to take the connection string from os.Getenv()
	conn, err := pgx.Connect(ctx, "postgresql://invent:PurpleOctopi*22@localhost:5432/inventory_service?sslmode=disable")
	failOnError(err, "Failed to connect to DB")
	defer conn.Close(ctx)

	var jsonOutput string
	err = conn.QueryRow(ctx, `
    		SELECT COALESCE(json_agg(t), '[]'::json)
    		FROM (
        		SELECT *
        		FROM inventory_events
        		WHERE product_id = $1
        		LIMIT $2
    		) t
	`, id, limit).Scan(&jsonOutput)
	failOnError(err, "Failed to query DB")

	return jsonOutput
}
