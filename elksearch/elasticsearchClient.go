package elksearch

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func InitElasticClient() (*elasticsearch.Client, error) {
	config, err := InitElasticConfig()
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	res, errr := Search(es, "umj3", "WOT")
	if errr != nil {
		fmt.Println(errr)
	}

	fmt.Println(res)

	return es, nil

}
