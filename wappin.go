package wappin

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"
)

const SendHsmEndpoint = "/v1/message/do-send-hsm"

type Config struct {
	ProjectId string
	SecretKey string
	ClientKey string
}

type Sender struct {
	Config      Config
	AccessToken AccessToken
}

type Wappin interface {
	postToWappin(endpoint string, payload interface{}) (ResMessage, error)
}

// Response body after post request to Wappin
type ResMessage struct {
	MessageId string `json:"message_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
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
	return &Sender{Config: config}
}

// Set authorization token
func (s *Sender) setToken() error {
	accessToken, err := getAccessToken(s.Config.SecretKey)

	if err != nil {
		return err
	}

	s.AccessToken = accessToken

	return nil
}

func (s *Sender) SendMessage(reqMsg interface{}) (res ResMessage, err error) {
	err = s.setToken()

	if err != nil {
		return ResMessage{
			MessageId: "",
			Status:    "400",
			Message:   err.Error(),
		}, err
	}

	switch req := reqMsg.(type) {
	case ReqWaMessage:
		res, err = s.sendWaMessage(req)
	default:
		return ResMessage{}, errors.New("invalid request message format")
	}

	return res, err
}

// Send Whatsapp message
func (s *Sender) sendWaMessage(req ReqWaMessage) (ResMessage, error) {
	res, err := s.postToWappin(SendHsmEndpoint, req)

	return res, err
}

// Post request to Wappin service
func (s *Sender) postToWappin(endpoint string, body interface{}) (ResMessage, error) {
	url := baseUrl + endpoint
	res, err := client.R().SetBody(body).SetAuthToken(s.AccessToken.Data.AccessToken).Post(url)
	resMessage := ResMessage{}

	if err != nil {
		log.Error(err)
		return resMessage, err
	}

	if res.StatusCode() != 200 {
	    log.WithFields(log.Fields{
			"msg": "Status code is not we expected",
			"res": res,
		}).Warn()
	}

	if err := json.Unmarshal(res.Body(), &resMessage); err != nil {
		return resMessage, err
	}

	return resMessage, err
}

