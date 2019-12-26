package wappin

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
		httpmock.ActivateNonDefault(client.GetClient())
		defer httpmock.DeactivateAndReset()

		fixture := `{ "status": "200", "message": "Success", "data": { "access_token": "677b800f9b694f98bb9db6edb18336743a3f416cadff1953a59190f309220936", "expired_datetime": "2020-12-28 10:20:23", "token_type": "Bearer" } }`
		responder := httpmock.NewStringResponder(200, fixture)
		fakeUrl := baseUrl + TokenEndpoint
		httpmock.RegisterResponder("POST", fakeUrl, responder)

		accessToken := getAccessToken("secret-key")

		assert.Equal(t, "677b800f9b694f98bb9db6edb18336743a3f416cadff1953a59190f309220936", accessToken.Data.AccessToken)
}
