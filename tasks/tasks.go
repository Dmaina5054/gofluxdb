package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	var p FluxdbFetchPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("Json failed to unmarshall: %v: %w", err, asynq.SkipRetry)

	}
	fmt.Println("Sent task download...")
	//code to fetch flux records here
	return nil

}
