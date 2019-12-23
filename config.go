package wappin

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	cacheDriver = os.Getenv("WAPPIN_CACHE_DRIVER")
	cacheHost   = os.Getenv("WAPPIN_CACHE_HOST")
	cachePort   = os.Getenv("WAPPIN_CACHE_PORT")
)
