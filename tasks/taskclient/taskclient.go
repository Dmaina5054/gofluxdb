package taskclient

import (
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
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

func ExecuteClient() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	//creating a task with typename and payload
	payload, err := json.Marshal(FluxdbFetchPayload{BucketName: "MWKs", DestinationBucket: "MWKsDownsampled"})
	if err != nil {
		log.Fatal(err)
	}

	//defining tasks
	t1 := asynq.NewTask("fluxdb:fetchrecords", payload)

	//try to start task immdiate
	info, err := client.Enqueue(t1)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(" [*] Completed running task: %+v", info)
}
