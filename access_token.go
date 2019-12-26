package wappin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/eko/gocache/store"
	"github.com/go-redis/redis/v7"
	"github.com/go-resty/resty/v2"
)

const TokenEndpoint = "/v1/token/get"

type AccessToken struct {
	ClientId string
	Status   string `json:"status"`
	Message  string `json:"message"`
	Data     struct {
		AccessToken     string `json:"access_token"`
		ExpiredDatetime string `json:"expired_datetime"`
		TokenType       string `json:"token_type"`
	} `json:"data"`
}

type keyTokenCache struct {
	ClientSecret string
}

var accessToken AccessToken
var cacheManager *cache.Cache
var marshal *marshaler.Marshaler
var client = resty.New()

func init() {
	initCacheManager()
}

func initCacheManager() {
	// TODO: Add dynamic options for driver
	store := store.NewRedis(redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cacheHost, cachePort),
	}), nil)

	cacheManager = cache.New(store)
	marshal = marshaler.New(cacheManager)
}

// Get access token from cache or calling API
func getAccessToken(clientSecret string) AccessToken {
	key := keyTokenCache{ClientSecret: clientSecret}
	value, err := marshal.Get(key, new(AccessToken))

	if value == nil && err != nil {
		accessToken = generateAccessToken(clientSecret)

		return accessToken
	}

	jsonBlob, err := json.Marshal(value)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonBlob, &accessToken)

	if err != nil {
		panic(err)
	}

	return accessToken
}

// Generate access token by calling token API endpoint
func generateAccessToken(clientSecret string) AccessToken {
	url := baseUrl + TokenEndpoint
	accessToken := AccessToken{}
	res, err := client.R().SetHeader("Authorization", getBasicAuth(clientSecret)).Post(url)

	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(res.Body(), &accessToken); err != nil {
		panic(err)
	}

	// Set cache
	setAccessToken(clientSecret, &accessToken)

	return accessToken
}

// Set access token in cache
func setAccessToken(clientSecret string, accessToken *AccessToken) {
	key := keyTokenCache{ClientSecret: clientSecret}
	seconds := expiredInSeconds(accessToken.Data.ExpiredDatetime)
	err := marshal.Set(key, accessToken, &store.Options{Tags: []string{"access_token"}, Expiration: seconds})

	if err != nil {
		panic(err)
	}
}

// Get expired time in seconds format
func expiredInSeconds(datetime string) time.Duration {
	layout := "2006-01-02 15:04:05 -07"
	expired, _ := time.Parse(layout, fmt.Sprintf("%s +07", datetime))
	duration := expired.Sub(time.Now())

	return time.Duration(int64(duration.Seconds())) * time.Second
}
