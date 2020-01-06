package wappin

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	baseUrl = os.Getenv("WAPPIN_BASE_URL")
	clientId = os.Getenv("WAPPIN_CLIENT_ID")
	cacheDriver = os.Getenv("WAPPIN_CACHE_DRIVER")
	cacheHost   = os.Getenv("WAPPIN_CACHE_HOST")
	cachePort   = os.Getenv("WAPPIN_CACHE_PORT")
)

