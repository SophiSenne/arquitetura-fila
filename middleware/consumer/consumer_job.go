package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"middleware/data"
	"middleware/model"
)

const (
	queueName     = "SENSOR.DATA"
	prefetchCount = 10          
	reconnectWait = 5 * time.Second
)

type Job struct {
	repo    *data.SensorRepository
	connStr string
	log     *slog.Logger
}

func NewJob(repo *data.SensorRepository, connStr string, log *slog.Logger) *Job {
	return &Job{repo: repo, connStr: connStr, log: log}
}

func (j *Job) Run(ctx context.Context) {
	j.log.Info("consumer job starting")

	for {
		if err := j.runOnce(ctx); err != nil {
			j.log.Error("consumer disconnected", "error", err)
		}

		select {
		case <-ctx.Done():
			j.log.Info("consumer job stopped")
			return
		case <-time.After(reconnectWait):
			j.log.Info("reconnecting to RabbitMQ...", "wait", reconnectWait)
		}
	}
}

func (j *Job) runOnce(ctx context.Context) error {
	conn, err := amqp.Dial(j.connStr)
	if err != nil {
		return fmt.Errorf("dialing RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("opening channel: %w", err)
	}
	defer ch.Close()

	if err = ch.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("setting QoS: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declaring queue %q: %w", queueName, err)
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("registering consumer: %w", err)
	}

	j.log.Info("listening", "queue", queueName)

	connClose := conn.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case <-ctx.Done():
			return nil

		case amqpErr, ok := <-connClose:
			if !ok {
				return fmt.Errorf("connection closed unexpectedly")
			}
			return fmt.Errorf("connection error: %s", amqpErr)

		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("deliveries channel closed")
			}
			j.handle(ctx, d)
		}
	}
}

func (j *Job) handle(ctx context.Context, d amqp.Delivery) {
	var msg model.SensorMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		j.log.Error("invalid message payload, discarding",
			"error", err,
			"body", string(d.Body),
		)
		_ = d.Nack(false, false)
		return
	}

	if err := j.repo.SaveReading(ctx, msg); err != nil {
		j.log.Error("failed to save reading, requeueing",
			"error", err,
			"idSensor", msg.IDSensor,
			"sensorType", msg.SensorType,
		)
		_ = d.Nack(false, true)
		return
	}

	j.log.Debug("reading saved",
		"idSensor", msg.IDSensor,
		"sensorType", msg.SensorType,
		"readType", msg.ReadType,
		"value", msg.Value,
	)
	_ = d.Ack(false)
}