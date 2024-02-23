package elksearch

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func InitElasticClient() {
	config, err := InitElasticConfig()
	if err !=nil{
		log.Fatal(err)
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil{
		log.Fatal(err)
	}

	log.Println(es.Info)


}
