package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"sensor-api/messaging"
	"sensor-api/model"
)

func Running(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": "API running"})
}

func ReceiveData(c *gin.Context) {
	var msg model.SensorMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Printf("[ReceiveData] Falha ao fazer parse do JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Printf("[ReceiveData] Mensagem recebida: %+v", msg)

	if err := messaging.ProduceMessage(msg); err != nil {
		log.Printf("[ReceiveData] Falha ao enviar mensagem para a fila: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ReceiveData] Mensagem enviada com sucesso para a fila")
	c.JSON(http.StatusOK, gin.H{"success": "Message sent to queue"})
}