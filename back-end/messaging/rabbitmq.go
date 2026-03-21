package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sensor-api/model"
)

const queueName = "SENSOR.DATA"

func ProduceMessage(message model.SensorMessage) error {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")

	if user == "" || password == "" || host == "" || port == "" {
		return fmt.Errorf("variáveis de ambiente do RabbitMQ não estão todas definidas (USER=%q, HOST=%q, PORT=%q, PASSWORD vazia=%v)",
			user, host, port, password == "")
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	log.Printf("[ProduceMessage] Conectando ao RabbitMQ: amqp://%s:***@%s:%s/", user, host, port)

	conn, err := connectRabbit(url)
	if err != nil {
		return fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}
	defer conn.Close()
	log.Printf("[ProduceMessage] Conexão estabelecida com sucesso")

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("erro ao abrir channel: %w", err)
	}
	defer ch.Close()
	log.Printf("[ProduceMessage] Channel aberto com sucesso")

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("erro ao declarar a fila: %w", err)
	}
	log.Printf("[ProduceMessage] Fila declarada: nome=%q | mensagens=%d | consumers=%d",
		q.Name, q.Messages, q.Consumers)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem para JSON: %w", err)
	}
	log.Printf("[ProduceMessage] Payload serializado: %s", string(jsonData))

	if err := ch.PublishWithContext(
		context.Background(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	); err != nil {
		return fmt.Errorf("erro ao publicar mensagem na fila %q: %w", q.Name, err)
	}

	log.Printf("[ProduceMessage] Mensagem publicada com sucesso na fila %q", q.Name)
	return nil
}

func connectRabbit(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		log.Printf("RabbitMQ não pronto, tentando novamente em 3s...")
		time.Sleep(3 * time.Second)
	}

	return nil, err
}