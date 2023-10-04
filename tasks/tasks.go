package tasks

import (
	"context"
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
	client.Options().SetHTTPRequestTimeout(uint(30000 * time.Second)).SetLogLevel(3)
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
