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
	fmt.Println(req.URL)

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
		//just to be safe despite filter in fluxquery
		if serialNo != "<nil>" {
			apires := enrichResult(serialNo.(string), apiSuffix, destBucket)

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
	prefixes := []string{"mwkn", "mwks", "stn", "kwd", "ksn", "krbs"}

	//iterate and check if exist
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowercaseInput, prefix) {
			fmt.Println(prefix)
			return prefix

		}

	}

	return ""

}
