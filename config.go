package wappin

import (
	"github.com/joho/godotenv"
	"os"
)

var (
	baseUrl string
	clientId string
)

func init() {
	loadEnv()

	baseUrl = os.Getenv("WAPPIN_BASE_URL")
	clientId = os.Getenv("WAPPIN_CLIENT_ID")
}


func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		godotenv.Load("./../../.env")
	}
}
