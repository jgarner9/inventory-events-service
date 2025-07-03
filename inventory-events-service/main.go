package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	conn := ConnectToRabbitMQ()
	if conn == nil {
		log.Panic("Failed to connect to RabbitMQ")
	}
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
	lastKnownQuantities["8c138fa0-bfb4-4fd3-a23a-fed6468337d3"] = 1
	lastKnownQuantities["a4a00466-c889-4bec-9eb2-89fb4950da6c"] = 0
	lastKnownQuantities["b1c31e32-721b-472f-84bc-35a95d595184"] = 5
	lastKnownQuantities["119eefcc-0477-47d8-9f4b-6c55c030a78b"] = 20

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

		lastKnownQuantities[productID] = currentQuantity

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

func ConnectToRabbitMQ() *amqp.Connection {
	var conn *amqp.Connection
	var err error

	for i := range 10 {
		conn, err = amqp.Dial(os.Getenv("RABBITMQ_URL"))
		if err == nil {
			log.Println("Connected to RabbitMQ")
			return conn
		}

		log.Printf("RabbitMQ not ready yet (%d/10): %s", i+1, err)
		time.Sleep(time.Duration(2*i+1) * time.Second) // exponential-ish backoff
	}

	log.Fatal("Could not connect to RabbitMQ after retries:", err)
	return nil
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
	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
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

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"test": "test",
	}))

	authorized.GET("/events", func(ctx *gin.Context) {
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
	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
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
