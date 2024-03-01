package elksearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
)

func SearchClient() (*elasticsearch.Client, error) {
	es, err := InitElasticClient()
	if err != nil {
		log.Printf("%w", err)
		return nil, err
	}
	return es, nil

}

func Search(client *elasticsearch.Client, indexName, serialCode string) ([]map[string]interface{}, error) {

	// Prepare the query DSL with the provided building
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"fuzzy": map[string]interface{}{
							"Serial_Code": serialCode,
						},
					},
				},
			},
		},
	}

	// Marshal the query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	// Prepare a bytes.Reader with the query JSON
	queryReader := bytes.NewReader(queryJSON)

	// Prepare the search request
	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  queryReader,
	}

	// Perform the search request
	res, err := req.Do(context.Background(), client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode the response
	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// Extract hits from the response
	hits, ok := response["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("hits not found in response")
	}

	// Extract _source from each hit
	var sources []map[string]interface{}
	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("hits array not found in response")
	}
	for _, hit := range hitsArray {
		source, ok := hit.(map[string]interface{})["_source"].(map[string]interface{})
		if !ok {
			continue
		}
		sources = append(sources, source)
	}

	return sources, nil
}
