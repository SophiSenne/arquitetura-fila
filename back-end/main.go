package main
 
import (
	"log"
 
	"sensor-api/router"
)
 
func main() {
	r := router.Setup()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}