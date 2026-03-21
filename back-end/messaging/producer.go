package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"sensor-api/model"
)

const queueName = "SENSOR.DATA"

// ProduceMessage publica uma mensagem na fila SENSOR.DATA reutilizando
// um channel do pool — sem abrir nova conexão por requisição.
func ProduceMessage(message model.SensorMessage) error {
	pool, err := GetPool()
	if err != nil {
		return fmt.Errorf("erro ao obter pool AMQP: %w", err)
	}

	entry, err := pool.Acquire()
	if err != nil {
		return fmt.Errorf("erro ao adquirir channel do pool: %w", err)
	}
	defer pool.Release(entry)

	q, err := entry.ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		// Marca o channel como inválido para que seja recriado na próxima Acquire.
		entry.broken = true
		return fmt.Errorf("erro ao declarar fila: %w", err)
	}

	log.Printf("[ProduceMessage] Fila: %q | mensagens=%d | consumers=%d",
		q.Name, q.Messages, q.Consumers)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	err = entry.ch.PublishWithContext(
		context.Background(),
		"",     // exchange padrão
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonData,
			DeliveryMode: amqp.Persistent, // sobrevive a restart do broker
		},
	)
	if err != nil {
		entry.broken = true
		return fmt.Errorf("erro ao publicar na fila %q: %w", q.Name, err)
	}

	log.Printf("[ProduceMessage] Mensagem publicada em %q: %s", q.Name, jsonData)
	return nil
}