package models

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ResultURL string `json:"result"`
}

type BatchRequest struct {
	BatchURLs []BatchURLRequest
}
type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	BatchURLs []BatchURLResponse
}

type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURLsResponse struct {
	UserURLs []*UserURL
}

type UserURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
