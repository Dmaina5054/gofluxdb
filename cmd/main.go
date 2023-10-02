package main

import (
	"log"

	"github.com/Dmaina5054/gofluxdb/tasks"
	"github.com/Dmaina5054/gofluxdb/tasks/taskclient"
	"github.com/hibiken/asynq"
)

func main() {

	taskclient.ExecuteClient()

	//new server to start the workers
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			LogLevel: asynq.InfoLevel,
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
