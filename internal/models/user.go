package models

// User user
type User struct {
	ID      int `json:"id"`
	URLs    []*StorageURL
	Service *UserService
}

// UserService дополнительная информация по пользователю
type UserService struct {
	IsAuthenticated bool
	Token           string
}
