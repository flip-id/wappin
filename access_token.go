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
	log "github.com/sirupsen/logrus"
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
	SecretKey string
}

var accessToken AccessToken
var cacheManager *cache.Cache
var marshal *marshaler.Marshaler
var (
	timeout, _ = strconv.Atoi(os.Getenv("WAPPIN_TIMEOUT"))
	client     = resty.New().SetTimeout(time.Second * time.Duration(timeout))
)
var (
	cacheDriver   string
	cacheHost     string
	cachePort     string
	cacheUser     string
	cachePassword string
)

func init() {
	loadEnv()
	prepareVars()
	initCacheManager()
}

func prepareVars() {
	cacheDriver = os.Getenv("WAPPIN_CACHE_DRIVER")
	cacheHost = os.Getenv("WAPPIN_CACHE_HOST")
	cachePort = os.Getenv("WAPPIN_CACHE_PORT")
	cacheUser = os.Getenv("WAPPIN_CACHE_USER")
	cachePassword = os.Getenv("WAPPIN_CACHE_PASSWORD")
}

func initCacheManager() {
	// TODO: Add dynamic options for driver
	connUrl := fmt.Sprintf("redis://%s:%s@%s:%s/0", cacheUser, cachePassword, cacheHost, cachePort)
	opt, err := redis.ParseURL(connUrl)

	if err != nil {
		panic(err)
	}

	store := store.NewRedis(redis.NewClient(opt), nil)

	cacheManager = cache.New(store)
	marshal = marshaler.New(cacheManager)
}

// Set custom client. This is useful for testing
func SetClient(c *resty.Client) {
	client = c
}

// Get access token from cache or calling API
func getAccessToken(secretKey string) (AccessToken, error) {
	key := keyTokenCache{SecretKey: secretKey}
	value, err := marshal.Get(key, new(AccessToken))

	if value == nil && err != nil {
		accessToken, err = generateAccessToken(secretKey)

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
func generateAccessToken(secretKey string) (AccessToken, error) {
	url := baseUrl + TokenEndpoint
	accessToken := AccessToken{}
	res, err := client.R().SetBasicAuth(clientId, secretKey).Post(url)

	if err != nil {
		return accessToken, err
	}

	if err := json.Unmarshal(res.Body(), &accessToken); err != nil {
		return accessToken, err
	}

	if accessToken.Status != "200" {
		log.WithFields(log.Fields{
			"msg": "Failed to get token",
			"res": res,
		}).Error()
		return accessToken, errors.New(accessToken.Message)
	}

	// Set cache
	err = setAccessToken(secretKey, &accessToken)

	return accessToken, err
}

// Set access token in cache
func setAccessToken(clientSecret string, accessToken *AccessToken) error {
	key := keyTokenCache{SecretKey: clientSecret}
	seconds := expiredInSeconds(accessToken.Data.ExpiredDatetime)
	err := marshal.Set(key, accessToken, &store.Options{Tags: []string{"access_token"}, Expiration: seconds})

	if err != nil {
		return errors.New("can't connect to cache driver")
	}

	return err
}

// Get expired time in seconds format
func expiredInSeconds(datetime string) time.Duration {
	layout := "2006-01-02 15:04:05 -07"
	expired, _ := time.Parse(layout, fmt.Sprintf("%s +07", datetime))
	duration := expired.Sub(time.Now())

	return time.Duration(int64(duration.Seconds())) * time.Second
}
