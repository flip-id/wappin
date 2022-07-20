//go:build integration
// +build integration

package wappin

import (
	"context"
	"flag"
	"github.com/fairyhunter13/dotenv"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
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
			dotenv.WithPaths(".env"),
		)
		if err != nil {
			log.Fatalln(err)
		}

		c = New(
			WithClientID(os.Getenv("WAPPIN_CLIENT_ID")),
			WithProjectID(os.Getenv("WAPPIN_PROJECT_ID")),
			WithSecretKey(os.Getenv("WAPPIN_SECRET_KEY")),
			WithClientKey(os.Getenv("WAPPIN_CLIENT_KEY")),
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

func TestSendOTPWAMessage(t *testing.T) {
	ctx := context.Background()
	resp, err := c.SendMessage(ctx, &RequestWhatsappMessage{
		Type:            "otp_code_new",
		RecipientNumber: os.Getenv("PHONE_NUMBER"),
		Params: map[string]string{
			"1": "202404",
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "200", resp.Status)
	assert.Equal(t, "Success", resp.Message)
	assert.Equal(t, http.StatusOK, resp.HttpStatusCode)
	assert.NotEmpty(t, resp.RawData)
	assert.NotEmpty(t, resp.MessageID)
	assert.Empty(t, resp.Data)
}
