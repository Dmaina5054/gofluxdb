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
	
	var payload FluxdbFetchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("json payload as %v", payload.BucketName)
	
	if err := godotenv.Load(); err != nil {
		log.Fatalf("erro loading .env file: %v", err)
	}

	// get influxdb config properties
	influxUrl := os.Getenv("INFLUX_URL")
	influxToken := os.Getenv("INFLUX_TK")
	
	

	//create a client
	client := influxdb2.NewClientWithOptions(influxUrl, influxToken, influxdb2.DefaultOptions().SetMaxRetries(5))
	client.Options().SetHTTPRequestTimeout(uint(30000 * time.Second)).SetLogLevel(4)
	defer client.Close()

	

	//create from task payload bucketname
	if _, err := fluxdb.InitClient(client,payload.BucketName); err != nil {
		log.Fatal(err)
	}

	log.Println("Done processing...")
	return nil

}
