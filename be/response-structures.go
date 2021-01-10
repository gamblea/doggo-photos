package main

// ErrorBody holds a error message to send to user
type ErrorBody struct {
	Error string `json:"error"`
}

// LoginResponse holds loginKey and sent to user
// on login success
type LoginResponse struct {
	LoginKey string `json:"loginKey"`
}

// UserRequest returns a username of user
type UserRequest struct {
	Username string `json:"username"`
}

// FEPhoto transformed to json to be sent to front end to hold
// photo metadata
type FEPhoto struct {
	Src    string `json:"src"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Date   string `json:"date"`
}

// UserDataResponse holds a user's photo metadata
type UserDataResponse struct {
	Photos []FEPhoto `json:"photos"`
}
