package fluxdb

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redis/go-redis/v9"
)

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
  |> range(start: -40m)
  |> filter(fn: (r) => r["_measurement"] == "interface")
  |> filter(fn: (r) => r["_field"] == "ifOperStatus")
  |> filter(fn: (r) => r["serialNumber"] != "")
  |> filter(fn: (r) => r["_value"] == 2)
 
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
		//define a context
		ctx := context.Background()

		redclient := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       6,
		})
		defer redclient.Close()

		//set key-value pair with expiry as half hour
		result, err := redclient.SetNX(ctx, serialNumber.(string), serialNumber, 30*time.Minute).Result()

		if err != nil {
			fmt.Println(err)
			continue

		}

		if result {
			//key !exist and was set enrich
			//define destination bucket
			destBucket := bucket + "Downsampled"

			//determine endpoint suffix for KOMP API
			kompApiSuffix := formatApiPrefix(bucket)

			client.Options().SetHTTPRequestTimeout(uint(30 * time.Second))
			defer client.Close()

			//initialize write api
			writeApi := client.WriteAPIBlocking("techops_monitor", destBucket)
			seenSerialNumbers := make(map[string]struct{})

			for range record.Values() {
				serialNo := record.ValueByKey("serialNumber").(string)

				if _, seen := seenSerialNumbers[serialNo]; !seen {
					// If not seen, add it to the set and print the record
					seenSerialNumbers[serialNo] = struct{}{}

					apires := enrichResult(serialNo, kompApiSuffix, destBucket)

					p := influxdb2.NewPointWithMeasurement("interface")

					p.AddField("GponPort", apires.Port)
					p.AddTag("OnuCode", apires.OnuCode)
					p.AddTag("OnuSerialNumber", apires.SerialNumber)
					p.AddTag("BuildingName", apires.BuildingName)
					p.AddTag("olt", fmt.Sprintf("%v", apires.Olt))
					p.AddTag("BuildingCode", apires.BuildingCode)
					p.AddTag("ClientName", apires.ClientName)
					p.AddTag("ClientContact", fmt.Sprintf("%v", apires.ClientContact))
					p.AddTag("GponPort", fmt.Sprintf("%v", apires.Port))

					p.SetTime(time.Now())
					//write point to bucket now
					writeApi.WritePoint(context.Background(), p)

				}

			}

		}

	}

	//check if any error during flux query

	if res.Err() != nil {
		log.Fatalf("Error reading record %v", res.Err().Error())
	}
	return "ok", err
}
