package models

type DelTaskStatus string

const (
	REGISTERED DelTaskStatus = "REGISTERED"
	DONE       DelTaskStatus = "DONE"
)

type DelTask struct {
	URL    string
	UserID int
	Status DelTaskStatus
}
