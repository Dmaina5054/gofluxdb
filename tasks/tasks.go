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
	BucketName        string
	DestinationBucket string
}

//create task
//task has type and payload

func NewFluxdbFetch(bucketName string, destinationBucket string) (*asynq.Task, error) {
	payload, err := json.Marshal(FluxdbFetchPayload{BucketName: bucketName, DestinationBucket: destinationBucket})
	if err != nil {
		return nil, err
	}
	fmt.Println("Starting tasks...")
	//no error
	//return new asynq task
	return asynq.NewTask(TypeFluxdbFetch, payload, asynq.MaxRetry(5), asynq.Timeout(30*time.Minute)), nil

}

//func to handle task xxx

func HandleFluxdbFetch(ctx context.Context, t *asynq.Task) error {
	fmt.Println(t.Type())
	
	if err := godotenv.Load(); err != nil {
		log.Fatalf("erro loading .env file: %v", err)
	}

	// get influxdb config properties
	influxUrl := os.Getenv("INFLUX_URL")
	influxToken := os.Getenv("INFLUX_TK")
	
	

	//create a client
	client := influxdb2.NewClientWithOptions(influxUrl, influxToken, influxdb2.DefaultOptions().SetMaxRetries(5))
	client.Options().SetHTTPRequestTimeout(uint(30000 * time.Second))
	defer client.Close()

	buckets := []string{"MWKn", "MWKs", "KSNOnu", "KWDOnu", "STNOnu"}
	for _, buck := range buckets {
		_, err := fluxdb.InitClient(client, buck)
		if err != nil {
			return err
		}

	}
	log.Println("Done processing...")
	return nil

}
