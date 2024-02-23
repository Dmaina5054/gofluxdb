package fluxdb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Dmaina5054/gofluxdb/elksearch"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redis/go-redis/v9"
)

var (
	uniqueItems = make(map[string]bool)
	mu          sync.Mutex
)

//Invoke Flux Query

func InitClient(client influxdb2.Client, bucket string) (string, error) {

	//test elk
	elksearch.SearchClient()
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
		ifDesc := record.ValueByKey("ifDescr")
		agentHost := record.ValueByKey("agent_host")
		olt := extractOlt(agentHost.(string))
		//fmt.Println(agentHost)
		_, err := extractNumber(ifDesc.(string))
		if err != nil {
			fmt.Println(err)
		}

		// fmt.Println(fmt.Sprintf("%s : %s", ifDesc, oltNumber))

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

			for range record.Values() {
				mu.Lock()
				if _, ok := uniqueItems[serialNumber.(string)]; !ok {
					uniqueItems[serialNumber.(string)] = true
					mu.Unlock()

					// Send the serial number
					fmt.Println("Sending ", serialNumber)
					apires := enrichResult(serialNumber.(string), kompApiSuffix, destBucket)

					p := influxdb2.NewPointWithMeasurement("iface")

					p.AddField("GponPort", apires.Port)
					p.AddTag("OnuCode", apires.OnuCode)
					p.AddTag("OnuSerialNumber", apires.SerialCode)
					p.AddTag("BuildingName", apires.BuildingName)
					p.AddTag("olt", fmt.Sprintf("%v", olt))
					p.AddTag("BuildingCode", apires.BuildingCode)
					p.AddTag("ClientName", apires.ClientName)
					p.AddTag("ClientContact", fmt.Sprintf("%v", apires.ClientContact))

					p.SetTime(time.Now())

					// Write point to bucket now
					writeApi.WritePoint(context.Background(), p)

				} else {
					mu.Unlock()
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

func extractNumber(label string) (string, error) {
	// Split the label by "/"
	parts := strings.Split(label, "/")

	// Extract the number part
	numberStr := parts[1]

	// Convert the number string to an integer
	number := strings.Split(numberStr, ":")
	fmt.Println(number[0])
	fmt.Println(number[1])
	fmt.Println(label)

	return number[0], nil
}

func extractOlt(hostip string) string {

	lastOctet := strings.Split(hostip, ".")[len(strings.Split(hostip, "."))-1]
	return lastOctet

}
