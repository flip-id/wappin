//go:build integration
// +build integration

package v2

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/fairyhunter13/dotenv"
	"github.com/flip-id/wappin/storage"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strconv"
	"sync"
	"testing"
)

var (
	c    Client
	once sync.Once
)

func setupClient() {
	once.Do(func() {
		err := dotenv.Load2(
			dotenv.WithPaths("../.env"),
		)
		if err != nil {
			log.Fatalln(err)
		}

		// setup redis and getting interface storage
		storage := setUpRedis()

		c = New(
			WithBaseURL(os.Getenv("WAPPIN_V2_BASE_URL")),
			WithLoginURL(os.Getenv("WAPPIN_V2_LOGIN_URL")),
			WithMessagesURL(os.Getenv("WAPPIN_V2_MESSAGES_URL")),
			WithUsername(os.Getenv("WAPPIN_V2_USERNAME")),
			WithPassword(os.Getenv("WAPPIN_V2_PASSWORD")),
			WithNamespace(os.Getenv("WAPPIN_V2_NAMESPACE")),
			WithTokenCacheKey(os.Getenv("WAPPIN_V2_TOKEN_CACHE_KEY")),
			WithStorage(storage),
		)
	})
}

func getRedisClient() (*redis.Client, error) {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return nil, errors.New("failed convert string DB to integer")
	}

	redisOptions := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		DB:   db,
	}

	if os.Getenv("REDIS_PASSWORD") != "" {
		redisOptions.Password = os.Getenv("REDIS_PASSWORD")
	}

	return redis.NewClient(redisOptions), nil
}

func setUpRedis() storage.IRedisStorage {
	redisClient, err := getRedisClient()
	if err != nil {
		fmt.Println("error setup redis", err)
	}

	return storage.NewGoRedisV8(redisClient)
}

// Run integration tests.
// Notes: Run this test only on local, not on CI/CD.
func TestMain(m *testing.M) {
	flag.Parse()
	setupClient()

	os.Exit(m.Run())
}

func TestSendMessage(t *testing.T) {
	ctx := context.Background()
	req := RequestMessage{
		To:   os.Getenv("PHONE_NUMBER"),
		Type: messageTypeTemplate,
		Template: TemplateRequest{
			Name: "testing_webhook_marketing",
			Language: LanguageRequest{
				Policy: "deterministic",
				Code:   "id",
			},
			Namespace: os.Getenv("WAPPIN_V2_NAMESPACE"),
			Components: []ComponentRequest{
				{
					Type: componentTypeBody,
					Parameters: []ComponentParameterRequest{
						{
							Type: messageTypeText,
							Text: "hari ini",
						},
						{
							Type: messageTypeText,
							Text: "Rp. 999999999",
						},
					},
				},
			},
		},
	}

	resp, err := c.SendMessage(ctx, &req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Messages[0].Id)

	fmt.Println("Success sending message to Wappin with message ID", resp.Messages[0].Id)
}

func TestSendMessageWithImage(t *testing.T) {
	ctx := context.Background()
	req := RequestMessage{
		To:   os.Getenv("PHONE_NUMBER"),
		Type: messageTypeTemplate,
		Template: TemplateRequest{
			Name: "testing_webhook_with_image",
			Language: LanguageRequest{
				Policy: "deterministic",
				Code:   "id",
			},
			Namespace: os.Getenv("WAPPIN_V2_NAMESPACE"),
			Components: []ComponentRequest{
				{
					Type: componentTypeHeader,
					Parameters: []ComponentParameterRequest{
						{
							Type: messageTypeImage,
							Image: &MediaParameterRequest{
								Link: "https://storage.googleapis.com/flip-prod-assets/images/verif_email.png",
							},
						},
					},
				},
			},
		},
	}

	resp, err := c.SendMessage(ctx, &req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Messages[0].Id)

	fmt.Println("Success sending message to Wappin with message ID", resp.Messages[0].Id)
}
