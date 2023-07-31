//go:build integration
// +build integration

package v2

import (
	"context"
	"flag"
	"github.com/fairyhunter13/dotenv"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
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

		c = New(
			WithBaseURL(os.Getenv("WAPPIN_V2_BASE_URL")),
			WithLoginURL(os.Getenv("WAPPIN_V2_LOGIN_URL")),
			WithMessagesURL(os.Getenv("WAPPIN_V2_MESSAGES_URL")),
			WithUsername(os.Getenv("WAPPIN_V2_USERNAME")),
			WithPassword(os.Getenv("WAPPIN_V2_PASSWORD")),
			WithNamespace(os.Getenv("WAPPIN_V2_NAMESPACE")),
			WithTokenCacheKey(os.Getenv("WAPPIN_V2_TOKEN_CACHE_KEY")),
		)
	})
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
}
