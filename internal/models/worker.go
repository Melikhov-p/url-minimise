package models

type DelTaskStatus string

const (
	Registered DelTaskStatus = "Registered"
	Done       DelTaskStatus = "Done"
)

type DelTask struct {
	URL    string
	UserID int
	Status DelTaskStatus
}
