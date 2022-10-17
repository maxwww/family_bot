package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/maxwww/family_bot/units"
	"log"
)

var _ units.TaskService = (*TaskService)(nil)

type TaskService struct {
	db *DB
}

func NewTaskService(db *DB) *TaskService {
	return &TaskService{db}
}

func (us *TaskService) CreateTask(ctx context.Context, task *units.Task) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createTask(ctx, tx, task); err != nil {
		return err
	}

	return tx.Commit()
}

func createTask(ctx context.Context, tx *sqlx.Tx, task *units.Task) error {
	query := `
	INSERT INTO tasks (title, date, done, notifications)
	VALUES ($1, $2, $3, $4) RETURNING id;
	`
	var date interface{} = nil
	if task.Date.Valid && task.Date.String != "" {
		date = task.Date.String
	}
	args := []interface{}{task.Title, date, false, task.Notifications}
	err := tx.QueryRowxContext(ctx, query, args...).Scan(&task.ID)

	if err != nil {
		switch {
		default:
			return err
		}
	}

	return nil
}

func (us *TaskService) TaskByID(ctx context.Context, taskId uint) (*units.Task, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	task, err := findOneTask(ctx, tx, units.TaskFilter{Id: &taskId})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return task, nil
}

func (us *TaskService) Tasks(ctx context.Context, tf units.TaskFilter) ([]*units.Task, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	tasks, err := findTasks(ctx, tx, tf)

	if err != nil {
		return nil, err
	}

	return tasks, tx.Commit()

	return []*units.Task{}, nil
}

func (us *TaskService) UpdateTask(ctx context.Context, task *units.Task, patch units.TaskPatch) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	defer tx.Rollback()

	if err := updateTask(ctx, tx, task, patch); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	return nil
}

func (us *TaskService) CompleteTask(ctx context.Context, taskId int) (bool, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return false, units.ErrInternal
	}

	defer tx.Rollback()

	query := `
	UPDATE tasks 
	SET done = not done
	WHERE id = $1
	RETURNING done;`

	var done bool
	tx.QueryRowxContext(ctx, query, taskId).Scan(&done)

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return false, units.ErrInternal
	}

	return done, nil
}

func (us *TaskService) RemoveCompete(ctx context.Context) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	defer tx.Rollback()

	query := `
	DELETE FROM tasks 
	WHERE done  = true;`

	tx.QueryRowxContext(ctx, query)

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	return nil
}

func (us *TaskService) RemoveByID(ctx context.Context, taskId int) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	defer tx.Rollback()

	args := []interface{}{taskId}
	query := `
	DELETE FROM tasks 
	WHERE id  = $1;`

	tx.QueryRowxContext(ctx, query, args...)

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	return nil
}

func findOneTask(ctx context.Context, tx *sqlx.Tx, filter units.TaskFilter) (*units.Task, error) {
	us, err := findTasks(ctx, tx, filter)

	if err != nil {
		return nil, err
	} else if len(us) == 0 {
		return nil, units.ErrNotFound
	}

	return us[0], nil
}

func findTasks(ctx context.Context, tx *sqlx.Tx, filter units.TaskFilter) ([]*units.Task, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0

	if v := filter.Id; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("id = $%d", argPosition)), append(args, *v)
	}

	query := "SELECT * from tasks" + formatWhereClause(where) +
		" ORDER BY id ASC" + formatLimitOffset(filter.Limit, filter.Offset)

	tasks, err := queryTasks(ctx, tx, query, args...)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func queryTasks(ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]*units.Task, error) {
	tasks := make([]*units.Task, 0)

	if err := findMany(ctx, tx, &tasks, query, args...); err != nil {
		return tasks, err
	}

	return tasks, nil
}

func updateTask(ctx context.Context, tx *sqlx.Tx, task *units.Task, patch units.TaskPatch) error {
	if v := patch.Done; v != nil {
		task.Done = *v
	}
	if v := patch.Title; v != nil {
		task.Title = *v
	}
	var dateValue interface{} = task.Date
	if v := patch.Date; v != nil {
		task.Date = *v
		dateValue = *v
		if !task.Date.Valid || task.Date.String == "" {
			dateValue = nil
		}
	}
	if v := patch.Notifications; v != nil {
		task.Notifications = *v
	}

	args := []interface{}{
		task.Done,
		task.Title,
		dateValue,
		task.Notifications,
		task.ID,
	}

	query := `
	UPDATE tasks 
	SET done = $1, title = $2, date = $3, notifications = $4
	WHERE id = $5`

	tx.QueryRowxContext(ctx, query, args...)

	return nil
}
