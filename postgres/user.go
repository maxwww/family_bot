package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/maxwww/family_bot/units"
	"log"
)

var _ units.UserService = (*UserService)(nil)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db}
}

func (us *UserService) CreateUser(ctx context.Context, user *units.User) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createUser(ctx, tx, user); err != nil {
		return err
	}

	return tx.Commit()
}

func (us *UserService) UserByTelegramID(ctx context.Context, telegramId uint) (*units.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := findOneUser(ctx, tx, units.UserFilter{TelegramID: &telegramId})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) Users(ctx context.Context, uf units.UserFilter) ([]*units.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	users, err := findUsers(ctx, tx, uf)

	if err != nil {
		return nil, err
	}

	return users, tx.Commit()

	return []*units.User{}, nil
}

func (us *UserService) UpdateUser(ctx context.Context, user *units.User, patch units.UserPatch) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	defer tx.Rollback()

	if err := updateUser(ctx, tx, user, patch); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return units.ErrInternal
	}

	return nil
}

func createUser(ctx context.Context, tx *sqlx.Tx, user *units.User) error {
	query := `
	INSERT INTO users (telegram_id, first_name, last_name, user_name)
	VALUES ($1, $2, $3, $4) RETURNING id;
	`
	args := []interface{}{user.TelegramID, user.FirstName, user.LastName, user.UserName}
	err := tx.QueryRowxContext(ctx, query, args...).Scan(&user.ID)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_telegram_id_key"`:
			return units.ErrDuplicateID
		default:
			return err
		}
	}

	return nil
}

func findOneUser(ctx context.Context, tx *sqlx.Tx, filter units.UserFilter) (*units.User, error) {
	us, err := findUsers(ctx, tx, filter)

	if err != nil {
		return nil, err
	} else if len(us) == 0 {
		return nil, units.ErrNotFound
	}

	return us[0], nil
}

func findUsers(ctx context.Context, tx *sqlx.Tx, filter units.UserFilter) ([]*units.User, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0

	if v := filter.TelegramID; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("telegram_id = $%d", argPosition)), append(args, *v)
	}

	query := "SELECT * from users" + formatWhereClause(where) +
		" ORDER BY id ASC" + formatLimitOffset(filter.Limit, filter.Offset)

	users, err := queryUsers(ctx, tx, query, args...)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func queryUsers(ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]*units.User, error) {
	users := make([]*units.User, 0)

	if err := findMany(ctx, tx, &users, query, args...); err != nil {
		return users, err
	}

	return users, nil
}

func updateUser(ctx context.Context, tx *sqlx.Tx, user *units.User, patch units.UserPatch) error {
	if v := patch.FirstName; v != nil {
		user.FirstName = *v
	}

	if v := patch.LastName; v != nil {
		user.LastName = *v
	}

	if v := patch.UserName; v != nil {
		user.UserName = *v
	}

	if v := patch.Notifications; v != nil {
		user.Notifications = *v
	}

	args := []interface{}{
		user.FirstName,
		user.LastName,
		user.UserName,
		user.Notifications,
		user.ID,
	}

	query := `
	UPDATE users 
	SET first_name = $1, last_name = $2, user_name = $3, notifications = $4
	WHERE id = $5`

	tx.QueryRowxContext(ctx, query, args...)

	return nil
}
