package main

type createAccountBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type tokenLoginBody struct {
	LoginKey string `json:"loginKey"`
}
