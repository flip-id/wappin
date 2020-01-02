package wappin

// Handler interface for sending Whatsapp message
type WaHandler interface {
	SendWaMessage(reqMessage ReqWaMessage) (ResMessage, error)
}

// Request body for Whatsapp message
type ReqWaMessage struct {
	ClientId        string
	ProjectId       string
	Type            string
	RecipientNumber string
	Params          map[string]string
}

