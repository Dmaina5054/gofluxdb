package elksearch

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func InitElasticClient() (*elasticsearch.Client, error) {
	config, err := InitElasticConfig()
	if err != nil {
		log.Printf("%w", err)
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		log.Printf("%w", err)
	}

	return es, nil

}
