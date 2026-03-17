package main

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
)

func running(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{"success": "API running"})
	return 
}

func receiveData(c *gin.Context){
	var body struct {
		IDSensor int `json:"int"`
		Timestamp string `json:"string"`
		SensorType string `json:"string"`
		ReadType string `json:"string"`
		Value float64 `json:"int"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}


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