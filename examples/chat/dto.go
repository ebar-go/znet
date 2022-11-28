package main

type LoginRequest struct {
	Name string `json:"name"`
}
type LoginResponse struct {
	Reply string `json:"reply"`
}
