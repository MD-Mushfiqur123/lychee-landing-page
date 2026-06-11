package api

type ConversationRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
