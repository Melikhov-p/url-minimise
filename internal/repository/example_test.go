package repository

import "fmt"

func ExampleNewEmptyUser() {
	user := NewEmptyUser()

	user.ID = 1

	fmt.Println(user.ID)

	// Output:
	// 1
}
