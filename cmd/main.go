package main

import (
	"log"
	"os"
	"sync"
	"time"
	"github.com/hibiken/asynq"
	"github.com/Dmaina5054/gofluxdb/tasks"

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

	//creating a waitgroup
	var wg sync.WaitGroup

	// Initialize Goroutine for periodic code run
	wg.Add(5)
	go fluxdb.InitClient(client, &wg, "MWKn")
	go fluxdb.InitClient(client, &wg, "MWKs")
	go fluxdb.InitClient(client, &wg, "STNOnu")
	go fluxdb.InitClient(client, &wg, "KSNOnu")
	go fluxdb.InitClient(client, &wg, "KWDOnu")


	log.Println("Waiting to complete goroutines")
	//wait for all goroutines to end
	wg.Wait()
	log.Println("Done processing Buckets")

}
