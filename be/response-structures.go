package main

type errorBody struct {
	Error string `json:"error"`
}

type loginResponse struct {
	LoginKey string `json:"loginKey"`
}

type userRequest struct {
	Username string `json:"username"`
}
