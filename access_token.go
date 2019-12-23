package wappin

import (
	"fmt"
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
	ClientId string
}

var accessToken AccessToken
var cacheManager *cache.Cache
var marshal *marshaler.Marshaler

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
func getAccessToken(clientId string) interface{} {
	key := keyTokenCache{ClientId: clientId}
	accessToken, err := marshal.Get(key, new(AccessToken))

	if accessToken == nil && err != nil {
		accessToken = generateAccessToken(clientId)

		return accessToken
	}

	return accessToken
}

// Generate access token by calling token API endpoint
func generateAccessToken(clientId string) AccessToken {
	client := resty.New()
	url := baseUrl + TokenEndpoint
	accessToken := AccessToken{}
	_, err := client.R().SetHeader("Authorization", credentials.getBasicAuth(clientId)).SetResult(&accessToken).Post(url)

	if err != nil {
		panic(err)
	}

	// Set cache
	setAccessToken(clientId, accessToken)

	return accessToken
}

// Set access token in cache
func setAccessToken(clientId string, accessToken AccessToken) {
	key := keyTokenCache{ClientId: clientId}
	err := marshal.Set(key, accessToken, &store.Options{Tags: []string{"access_token"}})

	if err != nil {
		panic(err)
	}
}
