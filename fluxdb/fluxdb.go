package fluxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
)

type EndpointResult struct {
	ClientContact int    `json:"ClientContact"`
	ClientName    string `json:"ClientName"`
	Active        int    `json:"active"`
	Region        string `json:"Region"`
	OnuStatus     string `json:"OnuStatus"`
	BuildingName  string `json:"BuildingName"`
	BuildingCode  string `json:"BuildingCode"`
	SerialCode    string `json:"Serial_Code"`
	MacAddress    string `json:"MacAddress"`
	OnuCode       string `json:"OnuCode"`
	ConfirmedBy   string `json:"confirmed_by"`
	SerialNumber  string `json:"Serial_Number"`
	GponNo        int    `json:"gpon_no"`
	Port          int    `json:"port"`
	Olt           int    `json:"olt"`
}

//passing of parameters for
//bucket or desired timestamp


func InitClient(client influxdb2.Client, bucket string)(string, error) {

	//define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 45* time.Second)
	defer cancel()
	// define queryApi
	queryApi := client.QueryAPI("techops_monitor")

	// Flux query
	fluxQuery := fmt.Sprintf(`
	from(bucket: "%s")
  |> range(start: -30s)
  |> filter(fn: (r) => r["_measurement"] == "interface")
  |> filter(fn: (r) => r["_field"] == "ifOperStatus")
  |> filter(fn: (r) => r["_value"] == 2)
  |> distinct(column: "_table")
   |> distinct(column: "serialNumber")
  |> first()
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
		serialNumberValue, ok := serialNumber.(string)

		if !ok {
			log.Printf("Warning: serial number %v\n got expected string", serialNumber)
			continue

		}

		//fonnd serialNUmber, enrich
		log.Printf("Processing serialNumberValue: %s for Bucket %s", serialNumberValue, bucket)

		//define destination bucket
		destBcket := bucket + "Downsampled"

		//determine endpoint suffix for KOMP API
		kompApiSuffix := formatApiPrefix(bucket)

		performTransformation(record, client, destBcket, kompApiSuffix)

		time.Sleep(5 * time.Second)

	}

	//check if any error during flux query
	//response parsing

	if res.Err() != nil {
		log.Fatalf("Error reading record %v", res.Err().Error())
	}
return "ok", err
}

func enrichResult(serialNumber string, apiSuffix string, destBucket string) EndpointResult {

	// Define the API endpoint and parameters
	kompApi := os.Getenv("KOMP_API_URL")
	fullApiURL := kompApi + "/" + apiSuffix
	kompJwt := os.Getenv("KOMP_JWT")
	serialCode := serialNumber

	// Create an HTTP client
	client := &http.Client{
		Timeout: 100000 * time.Second,
	}

	// Create an HTTP GET request
	req, err := http.NewRequest("GET", fullApiURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)

	}

	// Set request headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", kompJwt)
	req.Header.Add("X-CSRF-TOKEN", "")

	// Add query parameters
	q := req.URL.Query()
	q.Add("serial_code", serialCode)
	req.URL.RawQuery = q.Encode()

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)

	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)

	}

	// Print the response body
	var result []EndpointResult

	//unmarshall
	if err := json.Unmarshal(body, &result); err != nil {
		log.SetPrefix("DEBUG: ")

		// Check the type of error using type switching
		switch e := err.(type) {
		case *json.SyntaxError:
			log.Printf("Syntax error in JSON: %s", e.Error())
		case *json.UnmarshalTypeError:
			log.Printf("Type error during JSON unmarshaling: %s", e.Error())
		default:
			log.Printf("Other JSON unmarshal error: %s", err.Error())
		}

	}
	if len(result) == 0 {

		//TODO: Write to base
		//write serialNumber without komp result to file
		logPath := os.Getenv("LOG_PATH")
		logFilename := logPath + "missedOnus.txt"
		//open for writing
		file, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)

		}
		defer file.Close()

		//only write unique
		uniquePayload := make(map[string]bool)

		writePayload := fmt.Sprint(serialCode + ":" + destBucket + "\n")
		//confirm unique before write
		if !uniquePayload[writePayload] {
			//perform write
			_, err = file.Write([]byte(writePayload))
			if err != nil {
				log.Println(err)
			}
			// Add the data to the set to mark it as written
			uniquePayload[writePayload] = true

		}

		log.Printf("Failed to get result for %s :", serialCode)
		return EndpointResult{}
	}

	return result[0]

}

// perform transformation
func performTransformation(record *query.FluxRecord, client influxdb2.Client, destBucket string, apiSuffix string) {

	client.Options().SetHTTPRequestTimeout(uint(30 * time.Second))
	defer client.Close()

	//initialize write api
	writeApi := client.WriteAPIBlocking("techops_monitor", destBucket)

	for range record.Values() {
		serialNo := record.ValueByKey("serialNumber")
		if serialNo != "<nil>" {
			apires := enrichResult(serialNo.(string), apiSuffix, destBucket)
			fmt.Println(apires.BuildingName)

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

// function to determine endpoint to be scrapped
func formatApiPrefix(bucketName string) string {
	lowercaseInput := strings.ToLower(bucketName)

	//define prefix handlers
	prefixes := []string{"mwkn", "mwks", "stn", "kwd", "ksn","krbs"}

	//iterate and check if exist
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowercaseInput, prefix) {
			fmt.Println(prefix)
			return prefix

		}

	}

	return ""

}
