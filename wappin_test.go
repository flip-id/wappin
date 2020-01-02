package wappin

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendNotificationHSM(t *testing.T) {
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	mockGetAccessToken()

	fixture := `{ "message_id": "id-123", "status": "200", "message": "Success" }`
	responder := httpmock.NewStringResponder(200, fixture)
	fakeUrl := baseUrl + SendHsmEndpoint
	httpmock.RegisterResponder("POST", fakeUrl, responder)

	config := Config{
		ProjectId:    "0123",
		ClientSecret: "cs-key",
		ClientKey:    "ck-key",
	}
	sender := New(config)
	reqMsg := ReqWaMessage{
		ClientId:        "123",
		ProjectId:       "0123",
		Type:            "template_name",
		RecipientNumber: "089891234123",
		Params: map[string]string{
			"1": "John",
			"2": "Depok",
		},
	}

	res, _ := sender.SendMessage(reqMsg)

	assert.Equal(t, "id-123", res.MessageId)
	assert.Equal(t, "200", res.Status)
	assert.Equal(t, "Success", res.Message)
}

func TestFailSendNotificationHSM(t *testing.T) {
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	mockGetAccessToken()

	fixture := `{ "message_id": "id-124", "status": "600", "message": "Not delivered, Contact validate Failed" }`
	responder := httpmock.NewStringResponder(200, fixture)
	fakeUrl := baseUrl + SendHsmEndpoint
	httpmock.RegisterResponder("POST", fakeUrl, responder)

	config := Config{
		ProjectId:    "0123",
		ClientSecret: "cs-key",
		ClientKey:    "ck-key",
	}
	sender := New(config)
	reqMsg := ReqWaMessage{
		ClientId:        "123",
		ProjectId:       "0123",
		Type:            "template_name",
		RecipientNumber: "089891234123",
		Params: map[string]string{
			"1": "John",
			"2": "Depok",
		},
	}

	res, _ := sender.SendMessage(reqMsg)

	assert.Equal(t, "id-124", res.MessageId)
	assert.Equal(t, "600", res.Status)
	assert.Equal(t, "Not delivered, Contact validate Failed", res.Message)
}

func TestInvalidRequestFormat(t *testing.T) {
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	mockGetAccessToken()

	config := Config{
		ProjectId:    "0123",
		ClientSecret: "cs-key",
		ClientKey:    "ck-key",
	}
	sender := New(config)
	var reqMsg  interface{}

	_, err := sender.SendMessage(reqMsg)

	assert.Equal(t, "invalid request message format", err.Error())
}

func mockGetAccessToken() {
	fixture := `{ "status": "200", "message": "Success", "data": { "access_token": "677b800f9b694f98bb9db6edb18336743a3f416cadff1953a59190f309220936", "expired_datetime": "2020-12-28 10:20:23", "token_type": "Bearer" } }`
	responder := httpmock.NewStringResponder(200, fixture)
	fakeUrl := baseUrl + TokenEndpoint
	httpmock.RegisterResponder("POST", fakeUrl, responder)
}
