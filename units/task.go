package units

import (
	"context"
	"database/sql"
)

type Task struct {
	ID            uint
	Title         string         `db:"title"`
	Date          sql.NullString `db:"date"`
	Done          bool           `db:"done"`
	Notifications int            `db:"notifications"`
}

type TaskPatch struct {
	Title         *string
	Done          *bool
	Date          *sql.NullString
	Notifications *int
}

type TaskFilter struct {
	Id   *uint
	Done *bool

	Limit  int
	Offset int
}

type TaskService interface {
	CreateTask(context.Context, *Task) error

	TaskByID(context.Context, uint) (*Task, error)

	Tasks(context.Context, TaskFilter) ([]*Task, error)

	UpdateTask(context.Context, *Task, TaskPatch) error

	CompleteTask(context.Context, int) (bool, error)

	RemoveCompete(context.Context) error

	RemoveByID(context.Context, int) error
}
