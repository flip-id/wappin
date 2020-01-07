package wappin

import (
	"errors"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"testing"
)
func init() {
	err := godotenv.Load()

	if err != nil {
		godotenv.Load("./../../.env")
	}
}
func TestGetAccessToken(t *testing.T) {
		cacheManager.Clear()
		httpmock.ActivateNonDefault(client.GetClient())
		defer httpmock.DeactivateAndReset()

		fixture := `{ "status": "200", "message": "Success", "data": { "access_token": "677b800f9b694f98bb9db6edb18336743a3f416cadff1953a59190f309220936", "expired_datetime": "2020-12-28 10:20:23", "token_type": "Bearer" } }`
		responder := httpmock.NewStringResponder(200, fixture)
		fakeUrl := baseUrl + TokenEndpoint
		httpmock.RegisterResponder("POST", fakeUrl, responder)

		accessToken, _ := getAccessToken("secret-key")

		assert.Equal(t, "677b800f9b694f98bb9db6edb18336743a3f416cadff1953a59190f309220936", accessToken.Data.AccessToken)
}

func TestFailGetAccessToken(t *testing.T) {
	cacheManager.Clear()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	fixture := `{ "status": "401", "message": "Invalid credential", "data": null }`
	responder := httpmock.NewStringResponder(200, fixture)
	fakeUrl := baseUrl + TokenEndpoint
	httpmock.RegisterResponder("POST", fakeUrl, responder)

	_, err := getAccessToken("invalid-secret-key")

	expectedErr := errors.New("Invalid credential")
	assert.Equal(t, expectedErr, err)
}
