package wappin

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const SEND_HSM_ENDPOINT = "/v1/message/do-send-hsm"
const TOKEN_ENDPOINT = "/v1/token/get"
const CONNECTION_TIME_OUT = 15

var client = resty.New().SetTimeout(time.Second * time.Duration(CONNECTION_TIME_OUT))

type Config struct {
	BaseUrl   string
	ClientId  string
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
	MessageId      string `json:"message_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	HttpStatusCode int
	RawData        string
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

// Create sender object
func New(config Config) *Sender {
	return &Sender{Config: config}
}

func (s *Sender) SendMessage(reqMsg interface{}) (res ResMessage, err error) {

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
	res, err := s.postToWappin(SEND_HSM_ENDPOINT, req)

	return res, err
}

// Post request to Wappin service
func (s *Sender) postToWappin(endpoint string, body interface{}) (ResMessage, error) {
	url := s.Config.BaseUrl + endpoint
	res, err := client.R().SetBody(body).SetAuthToken(s.AccessToken.Data.AccessToken).Post(url)
	resMessage := ResMessage{}

	if res != nil {
		resMessage.HttpStatusCode = res.StatusCode()
		resMessage.RawData = string(res.Body())
	}

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

	return resMessage, nil
}

func (s *Sender) GenerateAccessToken() (AccessToken, error) {
	url := s.Config.BaseUrl + TOKEN_ENDPOINT
	accessToken := AccessToken{}
	res, err := client.R().SetBasicAuth(clientId, s.Config.SecretKey).Post(url)

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
	//err = setAccessToken(s.Config.SecretKey, &accessToken)

	return accessToken, err
}
