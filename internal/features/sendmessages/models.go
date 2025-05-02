package sendmessages

type WebhookPayload struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

type HookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}
