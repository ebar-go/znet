package main

type LoginRequest struct {
	Name string `json:"name"`
}
type LoginResponse struct {
	ID string `json:"id"`
}

type SendUserMessageRequest struct {
	ReceiverID string `json:"receiverId"`
	Content    string `json:"content"`
}
type SendUserMessageResponse struct {
	ID string `json:"id"`
}

type SubscribeChannelRequest struct{}
type SubscribeChannelResponse struct{}

type SendChannelMessageRequest struct{}
type SendChannelMessageResponse struct{}

type QueryHistoryMessageRequest struct{}
type QueryHistoryMessageResponse struct{}

type Message struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}
