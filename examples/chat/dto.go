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

type SubscribeChannelRequest struct {
	Name string `json:"name"`
}
type SubscribeChannelResponse struct{}

type SendChannelMessageRequest struct{}
type SendChannelMessageResponse struct{}

type QueryHistoryMessageRequest struct{}
type QueryHistoryMessageResponse struct{}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Sender  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"sender"`
	CreatedAt int64 `json:"createdAt"`
}
