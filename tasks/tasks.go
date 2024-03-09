package tasks

import (
	"context"
	"encoding/json"
	"log"
	"os"

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

	if err := godotenv.Load(); err != nil {
		log.Printf("error loading .env file: %v", err)
	}

	// get influxdb config properties
	influxUrl := os.Getenv("INFLUX_URL")
	//influxToken := os.Getenv("INFLUX_TK")
	influxOptions := influxdb2.DefaultOptions().SetHTTPRequestTimeout(10000).SetMaxRetries(10)

	client := influxdb2.NewClientWithOptions(influxUrl, "n_Vh5_hbrV8V5d2nppDcuhEsqy9UqOynuPuosuTlo0TS7YjYZbm4hA3MaSaFXGzuClHVJjwWOPJvNj5slcSe1Q==", influxOptions)

	defer client.Close()
	//create from task payload bucketname
	if _, err := fluxdb.InitClient(client, payload.BucketName); err != nil {
		//log.Printf("Err %w", err)
		log.Panic(err)
		recover()
	}

	log.Println("Done processing...")
	return nil

}
