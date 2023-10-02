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
//satisfies asynq.HandleFunc interface
//can also be a type that satisfies asynq.Handler interface

// need to accept, ctx and pointer to task
// need to return error if task failed
// use HandlerFunc adapter that
// allows ordinary funcs as handlers
func HandleFluxdbFetch(ctx context.Context, t *asynq.Task) error {
	fmt.Println(t.Type())
	//to store unmarshalled res
	// var p FluxdbFetchPayload
	// if err := json.Unmarshal(t.Payload(), &p); err != nil {
	// 	return fmt.Errorf("Jsoniled to unmarshall: %v: %w", err, asynq.SkipRetry)

	// }
	// fmt.Println("Sent task download...")

	//code to fetch flux records here
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

	buckets := []string{"MWKn", "MWKs", "STNOnu", "KSNOnu", "KWDOnu"}
	for _, buck := range buckets {
		_, err := fluxdb.InitClient(client, buck)
		if err != nil {
			return err
		}

	}
	log.Println("Done processing...")
	return nil

}
