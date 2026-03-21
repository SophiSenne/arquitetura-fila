package router

import (
	"github.com/gin-gonic/gin"

	"sensor-api/handler"
)

func Setup() *gin.Engine {
	r := gin.Default()

	r.GET("/", handler.Running)
	r.POST("/sensorData", handler.ReceiveData)

	return r
}