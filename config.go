package wappin

import (
	"github.com/joho/godotenv"
)

var (
	baseUrl string
	clientId string
)

func init() {
	loadEnv()

}

func NewConfig(bURL string, clID string){
	baseUrl = bURL
	clientId = clID
}

func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		godotenv.Load("./../../.env")
	}
}
