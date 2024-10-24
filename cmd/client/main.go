package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// MyAPIError — описание ошибки при неверном запросе.
type MyAPIError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	var users []User

	var respError MyAPIError

	client := resty.New()
	url := "https://jsonplaceholder.typicode.com/users"
	_, err := client.R().
		SetError(&respError).
		SetResult(&users).
		Get(url)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}
