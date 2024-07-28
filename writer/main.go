package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	System      string `json:"system"`
	EmployeeID  int    `json:"employee_id"`
	MessageBody string `json:"message_body"`
}

type AuditLog struct {
	Message   string    `bson:"message"`
	Timestamp time.Time `bson:"timestamp"`
	Status    string    `bson:"status"`
	RouterKey string    `bson:"router_key"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	initQueuesExchange(ch)

	routing := os.Args[1]
	message := os.Args[2]

	fmt.Fprintln(os.Stdout, routing)

	msg := Message{
		System:      "CRM",
		EmployeeID:  123,
		MessageBody: message,
	}

	body, err := json.Marshal(msg)
	failOnError(err, "Failed to marshal JSON")

	err = ch.Publish(
		"notify", // exchange
		routing,  // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)

	recordMessage(string(body), routing)
}

func initQueuesExchange(ch *amqp.Channel) {

	// Declarar a exchange
	err := ch.ExchangeDeclare(
		"notify", // nome da exchange
		"topic",  // tipo
		true,     // durável
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Falha ao declarar a exchange")

	queues := map[string]string{
		"sys1.message": "sys1.v1.messages",
		"sys2.message": "sys2.v1.messages",
	}

	for key, queue := range queues {
		_, err = ch.QueueDeclare(
			queue, // nome da fila
			true,  // durável
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		failOnError(err, "Falha ao declarar a fila "+queue)

		// Fazer o bind da fila sys1.v1.messages com a rota sys1.message
		err = ch.QueueBind(
			queue,    // nome da fila
			key,      // chave de roteamento
			"notify", // nome da exchange
			false,
			nil,
		)
		failOnError(err, "Falha ao fazer o bind da fila "+queue)
	}
}

func recordMessage(message string, routing string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Falha ao conectar ao MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	mongoCollection := mongoClient.Database("audit").Collection("logs")

	auditLog := AuditLog{
		Message:   message,
		Timestamp: time.Now(),
		Status:    "Enviado para RabbitMQ com sucesso",
		RouterKey: routing,
	}

	_, err = mongoCollection.InsertOne(ctx, auditLog)
	if err != nil {
		log.Fatalf("Falha ao inserir registro no MongoDB: %v", err)
	} else {
		fmt.Println("Mensagem enviada e registrada com sucesso!")
	}
}
