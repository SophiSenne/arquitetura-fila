package main
 
import (
	"log"
 
	"sensor-api/router"
	"sensor-api/messaging"
)
 
func main() {
	pool, err := messaging.GetPool()
	if err != nil{
		r := router.Setup()
		if err := r.Run(":8080"); err != nil {
			log.Fatal(err)
		}
		defer pool.Close()
	}
	
}