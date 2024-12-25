package models

type User struct {
	ID      int `json:"id"`
	URLs    []*StorageURL
	Service *UserService
}

type UserService struct {
	IsAuthenticated bool
	Token           string
}
