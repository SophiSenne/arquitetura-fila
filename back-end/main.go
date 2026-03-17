package main

import (
	"encoding/json"
	"log"
	"context"
	"net/http"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SensorMessage struct {
	IDSensor   int     `json:"idSensor"`
	Timestamp  string  `json:"timestamp"`
	SensorType string  `json:"sensorType"`
	ReadType   string  `json:"readType"`
	Value      float64 `json:"value"`
}

func running(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{"success": "API running"})
	return 
}

func receiveData(c *gin.Context){

	var msg SensorMessage

	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := ProduceMessage(msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send to queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Message sent to queue"})


}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", running)
	r.POST("/sensorData", receiveData)

	return r
}

func main() {
	r := setupRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func ProduceMessage(message SensorMessage) error {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("SENSOR.DATA", true, false, false, false, nil,)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), "", q.Name, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
}