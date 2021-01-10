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

type FEPhoto struct {
	Src    string `json:"src"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Date   string `json:"date"`
}

type userDataResponse struct {
	Photos []FEPhoto `json:"photos"`
}
