package elksearch

import (
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)


func InitElasticConfig() (elasticsearch.Config, error){

	cert, error := os.ReadFile(os.Getenv("ELK_CERT_PATH"))
	if error!= nil{
		return elasticsearch.Config{}, error
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELK_URL"),
		} ,
		Username: os.Getenv("ELK_USERNAME"),
		Password: os.Getenv("ELK_PASSWORD"),
		CACert: cert,
	}
	return esConfig, nil



}