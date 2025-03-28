package models

// Request модель запроса
type Request struct {
	URL string `json:"url"`
}

// Response модель ответа
type Response struct {
	ResultURL string `json:"result"`
}

// BatchRequest запрос пачки URL
type BatchRequest struct {
	BatchURLs []BatchURLRequest
}

// BatchURLRequest структура URL в пачке
type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse ответ создания пачки URL
type BatchResponse struct {
	BatchURLs []BatchURLResponse
}

// BatchURLResponse структура URL в пачке в ответе
type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse ответ URL созданные пользователем
type UserURLsResponse struct {
	UserURLs []*UserURL
}

// UserURL структура пользовательского URL
type UserURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
