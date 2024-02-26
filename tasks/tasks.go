package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Dmaina5054/gofluxdb/fluxdb"
	"github.com/hibiken/asynq"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
)

// define list of task types
const (
	TypeFluxdbFetch = "fluxdb:fetchrecords"
)

//define payloads for tasks

// for TypeFluxdbFetch
type FluxdbFetchPayload struct {
	BucketName        string `json:"bucketName"`        // Added struct tags for proper unmarshaling
	DestinationBucket string `json:"destinationBucket"` // Added struct tags for proper unmarshaling
}

//func to handle task xxx

func HandleFluxdbFetch(ctx context.Context, t *asynq.Task) error {

	//test elastic connection

	var payload FluxdbFetchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("Error %w", err)
	}
	fmt.Printf("json payload as %v", payload.BucketName)

	if err := godotenv.Load(); err != nil {
		log.Printf("erro loading .env file: %v", err)
	}

	// get influxdb config properties
	influxUrl := os.Getenv("INFLUX_URL")
	influxToken := os.Getenv("INFLUX_TK")

	client := influxdb2.NewClientWithOptions(influxUrl, influxToken, influxdb2.DefaultOptions().SetMaxRetries(5))
	client.Options().SetHTTPRequestTimeout(uint(60000 * time.Millisecond)).SetLogLevel(4) // Set timeout to 60 seconds
	defer client.Close()
	//create from task payload bucketname
	if _, err := fluxdb.InitClient(client, payload.BucketName); err != nil {
		log.Printf("Err %w", err)
	}

	log.Println("Done processing...")
	return nil

}
