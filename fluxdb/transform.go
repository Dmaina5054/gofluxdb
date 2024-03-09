package fluxdb

import (
	"strings"
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

type ElasticResult struct {
}

//add elastic search functionality

// function to determine endpoint to be scrapped
func formatApiPrefix(bucketName string) string {
	lowercaseInput := strings.ToLower(bucketName)

	//define prefix handlers
	prefixes := []string{"mwkn", "mwks", "stn", "kwd", "ksn", "krbs", "htr", "umj3", "lsm", "kibr"}

	//iterate and check if exist
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowercaseInput, prefix) {

			return prefix

		}

	}

	return ""

}
