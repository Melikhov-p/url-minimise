package models

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ResultURL string `json:"result"`
}
