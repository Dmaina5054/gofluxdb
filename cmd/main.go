package main

import (
	"log"
<<<<<<< HEAD
=======
	"os"
	"sync"
	"time"
	"github.com/hibiken/asynq"
	"github.com/Dmaina5054/gofluxdb/tasks"
>>>>>>> f29e0d118cd8616f50ad482416ef5bda78b5fa69

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

	// //initialize the fluxdb client

	// //load environment variables

	// if err := godotenv.Load(); err != nil {
	// 	log.Fatalf("erro loading .env file: %v", err)
	// }

	// // get influxdb config properties
	// influxUrl := os.Getenv("INFLUX_URL")
	// influxToken := os.Getenv("INFLUX_TK")

	// //create a client
	// client := influxdb2.NewClient(influxUrl, influxToken)
	// client.Options().SetHTTPRequestTimeout(uint(30 * time.Second))
	// defer client.Close()

	// //creating a waitgroup
	// var wg sync.WaitGroup

	// // Initialize Goroutine for periodic code run
	// wg.Add(5)
	// go fluxdb.InitClient(client, &wg, "MWKn")
	// go fluxdb.InitClient(client, &wg, "MWKs")
	// go fluxdb.InitClient(client, &wg, "STNOnu")
	// go fluxdb.InitClient(client, &wg, "KSNOnu")
	// go fluxdb.InitClient(client, &wg, "KWDOnu")

	// log.Println("Waiting to complete goroutines")
	// //wait for all goroutines to end
	// wg.Wait()
	// log.Println("Done processing Buckets")

}
