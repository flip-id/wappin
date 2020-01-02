package wappin

import (
	"encoding/json"
	"errors"
)

const SendHsmEndpoint  = "/v1/message/do-send-hsm"

type Config struct {
	ProjectId string
	ClientSecret string
	ClientKey string
}

// Handler interface for sending Whatsapp message
type WaHandler interface {
	SendWaMessage(reqMessage ReqWaMessage) (ResMessage, error)
}

type Sender struct {
	Config Config
	AccessToken AccessToken
}

// Request body for Whatsapp message
type ReqWaMessage struct {
	ClientId string
	ProjectId string
	Type string
	RecipientNumber string
	Params map[string]string
}

// Response body after sending message to Wappin
type ResMessage struct {
	MessageId string `json:"message_id"`
	Status string `json:"status"`
	Message string `json:"message"`
	//Data string
}

// Callback data from Wappin
type CallbackData struct {
	CallbackType   string `json:"callback_type"`
	ClientId       string `json:"client_id"`
	ClientName     string `json:"client_name"`
	Environment    string `json:"environment"`
	MessageContent string `json:"message_content"`
	MessageId      string `json:"message_id"`
	ProjectId      string `json:"project_id"`
	ProjectName    string `json:"project_name"`
	SenderNumber   string `json:"sender_number"`
	StatusMessages string `json:"status_messages"`
	Timestamp      string `json:"timestamp"`
}

// Create sender object
func New(config Config) *Sender {
	return &Sender{Config:config}
}

// Set authorization token
func (s *Sender) setToken() {
	accessToken, err := getAccessToken(s.Config.ClientSecret)

	if err != nil {
		panic(err)
	}

	s.AccessToken = accessToken
}

func (s *Sender) SendMessage(reqMsg interface{}) (res ResMessage, err error) {
	s.setToken()

	switch req := reqMsg.(type) {
	case ReqWaMessage:
		res, err = s.SendWaMessage(req)
	default:
		return ResMessage{}, errors.New("invalid request message format")
	}

	return res, err
}

// Send Whatsapp message
func (s *Sender) SendWaMessage(req ReqWaMessage) (ResMessage, error) {
	url := baseUrl + SendHsmEndpoint
	res, err := client.R().SetBody(req).SetHeader("Authorization", s.AccessToken.Data.AccessToken).Post(url)
	resMessage := ResMessage{}

	if err := json.Unmarshal(res.Body(), &resMessage); err != nil {
		return resMessage, err
	}

	return resMessage, err
}