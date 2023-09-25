package main

import (
	"log"
	"os"
	"time"

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

	//run in seperate goroutines
	go fluxdb.InitCLient(client, "MWKn")
	go fluxdb.InitCLient(client, "MWKs")
	go fluxdb.InitCLient(client, "STNOnu")
	go fluxdb.InitCLient(client, "KWDOnu")
	go fluxdb.InitCLient(client, "KSNOnu")

	//keep main running for other goroutines to execute
	select {}

}
