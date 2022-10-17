package state

import (
	"time"
)

const (
	STATUS_IDLE                 Status = "idle"
	STATUS_EDIT_TASK_WAIT_TITLE Status = "edit_task_wait_title"
	STATUS_EDIT_TASK_WAIT_DATE  Status = "edit_task_wait_date"
	STATUS_EDIT_TASK_WAIT_TIME  Status = "edit_task_wait_time"
	STATUS_ADD_TASK_PARSED      Status = "add_task_parsed"
	STATUS_ADD_TASK_WAIT_TITLE  Status = "add_task_wait_title"
	STATUS_ADD_TASK_WAIT_DATE   Status = "add_task_wait_date"
	STATUS_ADD_TASK_WAIT_TIME   Status = "add_task_wait_time"
)

type StateService struct {
	store map[int]State
}

type Status string

type Task struct {
	ID            int
	Title         string
	Notifications int
	Date          *time.Time
}

var _ StateServiceI = (*StateService)(nil)

type State struct {
	Status Status
	Task   Task
}

func NewStateService() *StateService {
	return &StateService{
		store: map[int]State{},
	}
}

type StateServiceI interface {
	GetUserState(userId int) *State
	SetUserState(userId int, state State)
}

func (rs *StateService) GetUserState(userId int) *State {
	state, ok := rs.store[userId]
	if !ok {
		rs.store[userId] = State{
			Status: STATUS_IDLE,
			Task:   Task{},
		}
	}

	state = rs.store[userId]

	return &state
}
func (rs *StateService) SetUserState(userId int, state State) {
	rs.store[userId] = state
}
