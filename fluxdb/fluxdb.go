package fluxdb

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/redis/go-redis/v9"
)

var (
	redclient    *redis.Client
	influxClient *influxdb2.Client
)

func initRedisClient() {
	ctx := context.Background()
	redclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       6,
	})
	defer redclient.Close()

	if err := redclient.Ping(ctx).Err(); err != nil {
		panic(err)
	}
}

//Invoke Flux Query

func InitClient(client influxdb2.Client, bucket string) (string, error) {

	//define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	// define queryApi
	queryApi := client.QueryAPI("techops_monitor")

	// Flux query
	fluxQuery := fmt.Sprintf(`
	from(bucket: "%s")
  |> range(start: -15m)
  |> filter(fn: (r) => r["_measurement"] == "interface")
  |> filter(fn: (r) => r["_field"] == "ifOperStatus")
  |> filter(fn: (r) => r["serialNumber"] != "")
  |> filter(fn: (r) => r["_value"] == 2)
  |> group(columns: ["serialNumber"])
  |> aggregateWindow(every: 5m, fn: last, createEmpty: false)
  |> distinct(column: "serialNumber")
  |> yield(name: "last")
`, bucket)

	res, err := queryApi.Query(ctx, fluxQuery)
	if err != nil {
		//hande InfluxDB-specific error
		log.Printf("Influx Error: %v", err)
	}

	// process record if no error
	for res.Next() {

		record := res.Record()
		serialNumber := record.ValueByKey("serialNumber")

		//write to cache for unique
		SetCache(serialNumber.(string), serialNumber.(string), bucket, res.Record())

	}

	//check if any error during flux query

	if res.Err() != nil {
		log.Fatalf("Error reading record %v", res.Err().Error())
	}
	return "ok", err
}

//set cache for uniqueness

func SetCache(serialKey string, serialNumber string, bucket string, record *query.FluxRecord) {

	//define a context
	ctx := context.Background()

	//set key-value pair with expiry as half hour
	result, err := redclient.SetNX(ctx, serialKey, serialNumber, 30*time.Minute).Result()

	if err != nil {
		fmt.Println(err)
		return

	}
	if result {
		//key !exist and was set enrich
		//define destination bucket
		destBcket := bucket + "Downsampled"

		//determine endpoint suffix for KOMP API
		kompApiSuffix := formatApiPrefix(bucket)

		performTransformation(record, *influxClient, destBcket, kompApiSuffix)
	

	}

}
