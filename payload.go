package wappin

import (
	"github.com/fairyhunter13/phone"
	"github.com/flip-id/valuefirst/manager"
)

const (
	// TokenBearer is the bearer token for Wappin.
	TokenBearer = "Bearer "
)

// ResponseMessage is a response message from the Wappin.
type ResponseMessage struct {
	MessageId      string                 `json:"message_id"`
	Status         string                 `json:"status"`
	Message        string                 `json:"message"`
	Data           map[string]interface{} `json:"data"`
	HttpStatusCode int                    `json:"-"`
	RawData        string                 `json:"-"`
}

// RequestWhatsappMessage is a request message for sending WhatsApp message.
type RequestWhatsappMessage struct {
	ClientID        string            `json:"client_id"`
	ProjectID       string            `json:"project_id"`
	Type            string            `json:"type"`
	RecipientNumber string            `json:"recipient_number"`
	Params          map[string]string `json:"params"`
	Token           string            `json:"-"`
}

// Default returns the default request for send Whatsapp message in Wappin.
func (r *RequestWhatsappMessage) Default() *RequestWhatsappMessage {
	r.RecipientNumber = phone.NormalizeID(r.RecipientNumber, 0)
	return r
}

// CallbackData is a callback data from Wappin.
type CallbackData struct {
	MessageID      string `json:"message_id"`
	ClientID       string `json:"client_id"`
	ClientName     string `json:"client_name"`
	ProjectID      string `json:"project_id"`
	ProjectName    string `json:"project_name"`
	StatusMessages string `json:"status_messages"`
	MessageContent string `json:"message_content"`
	Environment    string `json:"environment"`
	Timestamp      string `json:"timestamp"`
	SenderNumber   string `json:"sender_number"`
	CallbackType   string `json:"callback_type"`
}

// AccessToken is an access token from Wappin.
type AccessToken struct {
	ClientID string `json:"-"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Data     struct {
		AccessToken     string `json:"access_token"`
		ExpiredDatetime string `json:"expired_datetime"`
		TokenType       string `json:"token_type"`
	} `json:"data"`
}

func (a *AccessToken) ToResponseGenerateToken(statusCode int) (token manager.ResponseGenerateToken, err error) {
	err = getError(statusCode, a.Status, a.Message)
	if err != nil {
		return
	}

	token.Token = a.Data.AccessToken
	token.ExpiryDate = a.Data.ExpiredDatetime
	return
}
