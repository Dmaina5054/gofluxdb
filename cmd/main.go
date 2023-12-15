package main

import (
	"log"

	"github.com/Dmaina5054/gofluxdb/tasks"
	"github.com/hibiken/asynq"
)



//define payloads for tasks

// for TypeFluxdbFetch
type FluxdbFetchPayload struct {
	BucketName        string
	DestinationBucket string
}

func main() {

	//new server to start the workers
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{Concurrency: 10,
			
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			LogLevel: asynq.DebugLevel,
		},
	)

	// // defining mux server
	mux := asynq.NewServeMux()
	mux.HandleFunc("fluxdb:fetchrecords", tasks.HandleFluxdbFetch)

	// // TODO: Declare a normal func and use asynq.HandleFunc
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}

}
