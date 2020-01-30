package wappin

// Handler interface for sending Whatsapp message
type WaHandler interface {
	sendWaMessage(reqMessage ReqWaMessage) (ResMessage, error)
}

// Request body for Whatsapp message
type ReqWaMessage struct {
	ClientId        string            `json:"client_id"`
	ProjectId       string            `json:"project_id"`
	Type            string            `json:"type"`
	RecipientNumber string            `json:"recipient_number"`
	Params          map[string]string `json:"params"`
}
