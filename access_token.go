package wappin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
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
var (
	timeout, _ = strconv.Atoi(os.Getenv("WAPPIN_TIMEOUT"))
	client     = resty.New().SetTimeout(time.Second * time.Duration(timeout))
)

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
func getAccessToken(clientSecret string) (AccessToken, error) {
	key := keyTokenCache{ClientSecret: clientSecret}
	value, err := marshal.Get(key, new(AccessToken))

	if value == nil && err != nil {
		accessToken, err = generateAccessToken(clientSecret)

		return accessToken, err
	}

	jsonBlob, err := json.Marshal(value)

	if err != nil {
		return accessToken, err
	}

	err = json.Unmarshal(jsonBlob, &accessToken)

	if err != nil {
		return accessToken, err
	}

	return accessToken, err
}

// Generate access token by calling token API endpoint
func generateAccessToken(clientSecret string) (AccessToken, error) {
	url := baseUrl + TokenEndpoint
	accessToken := AccessToken{}
	res, err := client.R().SetBasicAuth(clientId, clientSecret).Post(url)

	if err != nil {
		return accessToken, err
	}

	if err := json.Unmarshal(res.Body(), &accessToken); err != nil {
		return accessToken, err
	}

	if accessToken.Status == "401" {
		return accessToken, errors.New(accessToken.Message)
	}

	// Set cache
	err = setAccessToken(clientSecret, &accessToken)

	return accessToken, err
}

// Set access token in cache
func setAccessToken(clientSecret string, accessToken *AccessToken) error {
	key := keyTokenCache{ClientSecret: clientSecret}
	seconds := expiredInSeconds(accessToken.Data.ExpiredDatetime)
	err := marshal.Set(key, accessToken, &store.Options{Tags: []string{"access_token"}, Expiration: seconds})

	return err
}

// Get expired time in seconds format
func expiredInSeconds(datetime string) time.Duration {
	layout := "2006-01-02 15:04:05 -07"
	expired, _ := time.Parse(layout, fmt.Sprintf("%s +07", datetime))
	duration := expired.Sub(time.Now())

	return time.Duration(int64(duration.Seconds())) * time.Second
}
