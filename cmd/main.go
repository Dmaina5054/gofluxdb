package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Dmaina5054/gofluxdb/tasks"
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

func initScheduler() {

	//init scheduler
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: ":6379"},
		&asynq.SchedulerOpts{Location: time.Local,
		LogLevel: asynq.DebugLevel,},
	)
	payload, err := json.Marshal(FluxdbFetchPayload{BucketName: "MWKs", DestinationBucket: "MWKsDownsampled"})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := scheduler.Register("*/5 * * * *", asynq.NewTask(TypeFluxdbFetch, payload)); err != nil {
		log.Fatal(err)
	}

	// Run blocks and waits for os signal to terminate the program.
	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}

}

func main() {

	//initScheduler()

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
