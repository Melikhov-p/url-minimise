package models

// DelTaskStatus тип для статуса задачи на удаление
type DelTaskStatus string

// Статусы задачи на удаление
const (
	Registered DelTaskStatus = "Registered"
	Done       DelTaskStatus = "Done"
)

// DelTask структура задачи на удаление URL
type DelTask struct {
	URL    string
	UserID int
	Status DelTaskStatus
}
