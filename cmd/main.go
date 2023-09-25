package main

import (
	"log"
	"os"
	"time"
	"sync"

	"github.com/Dmaina5054/gofluxdb/fluxdb"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
)

func main() {
	//initialize the fluxdb client

	//load environment variables

	if err := godotenv.Load(); err != nil {
		log.Fatalf("erro loading .env file: %v", err)
	}

	// get influxdb config properties
	influxUrl := os.Getenv("INFLUX_URL")
	influxToken := os.Getenv("INFLUX_TK")

	//create a client
	client := influxdb2.NewClient(influxUrl, influxToken)
	client.Options().SetHTTPRequestTimeout(uint(30 * time.Second))
	defer client.Close()

	//create a ticker to tick every 15 minutes
	ticker := time.NewTicker(15 * time.Minute)

	//creating a waitgroup
	var wg sync.WaitGroup

	// Initialize Goroutine for periodic code run
	
		for _, bucketName:= range []string{"MWKn", "MWKs", "STNOnu", "KWDOnu", "KSNOnu"} {

			wg.Add(1)
			go func(bucket string){
				defer wg.Done()
				for {
					fluxdb.InitCLient(client,bucket)

					//wait for next tick
					<-ticker.C
				}

			}(bucketName)
			
		}

		//wait for all goroutines to end
		wg.Wait()
	

	

}
