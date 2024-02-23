package elksearch

import (
	"fmt"
	"log"
)

func SearchClient() {
	es, err := InitElasticClient()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(es.Info())

}
