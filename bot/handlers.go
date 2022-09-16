package bot

import (
	"context"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	st "github.com/maxwww/family_bot/state"
	"github.com/maxwww/family_bot/units"
	"log"
	"time"
)

const (
	commandStart       = "start"
	commandHelp        = "help"
	commandList        = "list"
	commandCancel      = "cancel"
	commandSubscribe   = "subscribe"
	commandUnsubscribe = "unsubscribe"

	CQNewTaskSave          = "new_task_save"
	CQNewTaskEditTitle     = "new_task_edit_title"
	CQNewTaskEditDay       = "new_task_edit_day"
	CQNewTaskRemoveDay     = "new_task_remove_day"
	CQNewTaskEditTime      = "new_task_edit_time"
	CQNewTaskRemoveTime    = "new_task_remove_time"
	CQNewTaskCancel        = "new_task_cancel"
	CQTaskComplete         = "task_complete"
	CQTaskEdit             = "task_edit"
	CQTaskEditOk           = "edit_task_ok"
	CQTaskRemoveAllDone    = "task_remove_all_done"
	CQTaskRemoveAllDoneYes = "task_remove_all_done_yes"
	CQTaskRemoveAllDoneNo  = "task_remove_all_done_no"
	CQTaskEditEditDay      = "task_edit_edit_day"
	CQTaskEditRemoveDay    = "task_edit_remove_day"
	CQTaskEditEditTime     = "task_edit_edit_time"
	CQTaskEditRemoveTime   = "task_edit_remove_time"
	CQTaskEditEditTitle    = "task_edit_edit_title"
	CQTaskEditDeleteTask   = "task_edit_delete_task"
)

// command handlers
func (bot *Bot) handleStartCommand(chatId int64) {
	bot.sendMessage(chatId, TextStartMessage, nil)
}

func (bot *Bot) handleListCommand(chatId int64) {
	message, keyboard := bot.getTasksListWithHeader()

	bot.sendMessage(chatId, message, keyboard)
}

func (bot *Bot) handleCancelCommand(chatId int64, userTelegramId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})

	bot.sendMessage(chatId, TextCancel, nil)
}

func (bot *Bot) handleSubscribeCommand(chatId int64, notifications bool, user *units.User) {
	err := bot.userService.UpdateUser(context.Background(), user, units.UserPatch{
		Notifications: &notifications,
	})

	if err != nil {
		bot.sendGeneralError(chatId)
		return
	}

	message := TextSubscriptionsOn

	if !user.Notifications {
		message = TextSubscriptionsOff
	}

	bot.sendMessage(chatId, message, nil)
}

func (bot *Bot) handleUnknownCommand(chatId int64) {
	bot.sendMessage(chatId, TextUnknownCommand, nil)
}

// message handlers
func (bot *Bot) handleIdleMessage(chatId int64, user *units.User, message string) {
	date, title, _, err := bot.findDate(message)
	if err != nil {
		log.Println(err)
		bot.sendParseError(chatId)
		return
	}

	title = trim(title)

	if title != "" {
		ms := getNewTaskInfo(title, date)
		keyboard := buildEditTaskKeyboard(date, 0)

		bot.stateService.SetUserState(int(user.TelegramID), st.State{
			Status: st.STATUS_ADD_TASK_PARSED,
			Task: st.Task{
				Title: title,
				Date:  date,
			},
		})

		bot.sendMessage(chatId, ms, keyboard)
	} else {
		bot.sendParseError(chatId)
	}
}

func (bot *Bot) handleNewTaskEditTitle(chatId int64, user *units.User, task st.Task, message string) {
	title := trim(message)

	if title != "" {
		ms := getNewTaskInfo(title, task.Date)
		keyboard := buildEditTaskKeyboard(task.Date, 0)

		bot.stateService.SetUserState(int(user.TelegramID), st.State{
			Status: st.STATUS_ADD_TASK_PARSED,
			Task: st.Task{
				Title: title,
				Date:  task.Date,
			},
		})

		bot.sendMessage(chatId, ms, keyboard)
	} else {
		bot.sendParseError(chatId)
	}
}

func (bot *Bot) handleNewTaskEditDate(chatId int64, user *units.User, task st.Task, message string) {
	date, _, _, err := bot.findDate(message)
	if err != nil {
		log.Println(err)
		bot.sendParseError(chatId)
		return
	}

	if task.Date != nil && (task.Date.Hour() != 0 || task.Date.Minute() != 0) && date.Hour() == 0 && date.Minute() == 0 {
		*date = time.Date(date.Year(), date.Month(), date.Day(), task.Date.Hour(), task.Date.Minute(), task.Date.Second(), 0, bot.loc)
	}

	ms := getNewTaskInfo(task.Title, date)
	keyboard := buildEditTaskKeyboard(date, 0)

	bot.stateService.SetUserState(int(user.TelegramID), st.State{
		Status: st.STATUS_ADD_TASK_PARSED,
		Task: st.Task{
			Title: task.Title,
			Date:  date,
		},
	})

	bot.sendMessage(chatId, ms, keyboard)
}

func (bot *Bot) handleNewTaskEditTime(chatId int64, user *units.User, task st.Task, message string) {
	date, _, isDateFound, err := bot.findDate(message)
	if err != nil {
		log.Println(err)
		bot.sendParseError(chatId)
		return
	}

	var newDate *time.Time
	if isDateFound {
		newDate = date
	} else {
		tmp := time.Date(task.Date.Year(), task.Date.Month(), task.Date.Day(), date.Hour(), date.Minute(), date.Second(), 0, bot.loc)
		newDate = &tmp
	}

	ms := getNewTaskInfo(task.Title, newDate)
	keyboard := buildEditTaskKeyboard(newDate, 0)

	bot.stateService.SetUserState(int(user.TelegramID), st.State{
		Status: st.STATUS_ADD_TASK_PARSED,
		Task: st.Task{
			Title: task.Title,
			Date:  newDate,
		},
	})

	bot.sendMessage(chatId, ms, keyboard)
}

func (bot *Bot) handleEditTaskEditTitle(chatId int64, userTelegramId int, message string, taskId int) {
	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err != nil {
		log.Println(err)
		bot.sendParseError(chatId)
		return
	}

	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})

	title := trim(message)
	err = bot.taskService.UpdateTask(context.Background(), task, units.TaskPatch{
		Title: &title,
	})

	if err != nil {
		log.Println(err)
		bot.sendGeneralError(chatId)
		return
	}

	date := bot.getDateFromNullString(task.Date)
	ms := getEditingTaskInfo(task.Title, date)
	keyboard := buildEditTaskKeyboard(date, int(task.ID))

	bot.sendMessage(chatId, ms, keyboard)
}

func (bot *Bot) handleEditTaskEditDate(chatId int64, userTelegramId int, message string, taskId int) {
	date, _, isDateFound, err := bot.findDate(message)

	if err != nil || !isDateFound {
		bot.sendParseError(chatId)
		return
	}

	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})

	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err != nil || !isDateFound {
		bot.sendGeneralError(chatId)
		return
	}

	newDate := sql.NullString{}
	if isDateFound && date != nil {
		oldDate := bot.getDateFromNullString(task.Date)
		if oldDate != nil && (oldDate.Hour() != 0 || oldDate.Minute() != 0) && date.Hour() == 0 && date.Minute() == 0 {
			*date = time.Date(date.Year(), date.Month(), date.Day(), oldDate.Hour(), oldDate.Minute(), oldDate.Second(), 0, bot.loc)
		}
		newDate.String = date.Format(DateWithTimeFormat)
		newDate.Valid = true
	}

	err = bot.taskService.UpdateTask(context.Background(), task, units.TaskPatch{
		Date: &newDate,
	})
	if err != nil || !isDateFound {
		bot.sendGeneralError(chatId)
		return
	}

	ms := getEditingTaskInfo(task.Title, date)
	keyboard := buildEditTaskKeyboard(date, int(task.ID))

	bot.sendMessage(chatId, ms, keyboard)
}

func (bot *Bot) handleEditTaskEditTime(chatId int64, userTelegramId int, message string, taskId int) {
	date, _, isDateFound, err := bot.findDate(message)
	if err != nil {
		bot.sendParseError(chatId)
		return
	}

	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})

	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err != nil {
		bot.sendGeneralError(chatId)
		return
	}

	var newDate *time.Time
	if isDateFound {
		newDate = date
	} else {
		oldDate := bot.getDateFromString(task.Date.String)
		tmp := time.Date(oldDate.Year(), oldDate.Month(), oldDate.Day(), date.Hour(), date.Minute(), date.Second(), 0, bot.loc)
		newDate = &tmp
	}

	dateForUpdate := sql.NullString{
		String: newDate.Format(DateWithTimeFormat),
		Valid:  true,
	}
	err = bot.taskService.UpdateTask(context.Background(), task, units.TaskPatch{
		Date: &dateForUpdate,
	})
	if err != nil {
		bot.sendGeneralError(chatId)
		return
	}

	ms := getEditingTaskInfo(task.Title, newDate)
	keyboard := buildEditTaskKeyboard(newDate, int(task.ID))

	bot.sendMessage(chatId, ms, keyboard)
}

// callback handlers
func (bot *Bot) saveNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		newTask := createTaskStateWithDate(&state.Task)

		err := bot.taskService.CreateTask(context.Background(), newTask)
		if err != nil {
			bot.sendGeneralError(chatId)
			return
		}

		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_IDLE,
		})

		bot.deleteMessage(chatId, messageId)

		message := getSavedTaskInfo(state.Task.Title, state.Task.Date)

		for _, id := range bot.subscribers {
			user, err := bot.userService.UserByTelegramID(context.Background(), uint(id))
			if err == nil && user.Notifications {
				bot.sendMessage(id, message, nil)
			}
		}
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) editTitleNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_ADD_TASK_WAIT_TITLE,
			Task:   state.Task,
		})

		bot.deleteMessage(chatId, messageId)

		bot.sendMessage(chatId, TextSendNewTitle, nil)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) editDayNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_ADD_TASK_WAIT_DATE,
			Task:   state.Task,
		})

		bot.deleteMessage(chatId, messageId)

		bot.sendMessage(chatId, TextSendNewDay, nil)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) removeDayNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_ADD_TASK_PARSED,
			Task: st.Task{
				Title: state.Task.Title,
			},
		})

		message := getNewTaskInfo(state.Task.Title, nil)
		keyboard := buildEditTaskKeyboard(nil, 0)

		bot.editMessage(chatId, messageId, message, keyboard)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) editTimeNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_ADD_TASK_WAIT_TIME,
			Task:   state.Task,
		})

		bot.deleteMessage(chatId, messageId)

		bot.sendMessage(chatId, TextSendNewTime, nil)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) removeTimeNewTask(state *st.State, chatId int64, messageId int, userTelegramId int) {
	if state.Status == st.STATUS_ADD_TASK_PARSED {
		newDate := bot.getMidnightFromDate(state.Task.Date)

		bot.stateService.SetUserState(userTelegramId, st.State{
			Status: st.STATUS_ADD_TASK_PARSED,
			Task: st.Task{
				Title: state.Task.Title,
				Date:  newDate,
			},
		})

		message := getNewTaskInfo(state.Task.Title, newDate)
		keyboard := buildEditTaskKeyboard(newDate, 0)

		bot.editMessage(chatId, messageId, message, keyboard)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) cancelNewTask(chatId int64, messageId int, userTelegramId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})
	bot.deleteMessage(chatId, messageId)
}

func (bot *Bot) editTaskTitle(chatId int64, messageId int, userTelegramId int, taskId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_EDIT_TASK_WAIT_TITLE,
		Task: st.Task{
			ID: taskId,
		},
	})

	bot.editMessage(chatId, messageId, TextSendNewTitle, nil)
}

func (bot *Bot) editTaskDay(chatId int64, messageId int, userTelegramId int, taskId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_EDIT_TASK_WAIT_DATE,
		Task: st.Task{
			ID: taskId,
		},
	})

	bot.editMessage(chatId, messageId, TextSendNewDay, nil)
}

func (bot *Bot) editTaskTime(chatId int64, messageId int, userTelegramId int, taskId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_EDIT_TASK_WAIT_TIME,
		Task: st.Task{
			ID: taskId,
		},
	})

	bot.editMessage(chatId, messageId, TextSendNewTime, nil)
}

func (bot *Bot) removeTaskDay(chatId int64, messageId int, userTelegramId int, taskId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})
	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err != nil {
		log.Println(err)
		bot.sendGeneralError(chatId)
		return
	}

	newDate := sql.NullString{}
	err = bot.taskService.UpdateTask(context.Background(), task, units.TaskPatch{
		Date: &newDate,
	})
	if err != nil {
		log.Println(err)
		bot.sendGeneralError(chatId)
		return
	}

	message := getEditingTaskInfo(task.Title, nil)
	keyboard := buildEditTaskKeyboard(nil, taskId)

	bot.editMessage(chatId, messageId, message, keyboard)
}

func (bot *Bot) removeTaskTime(chatId int64, messageId int, userTelegramId int, taskId int) {
	bot.stateService.SetUserState(userTelegramId, st.State{
		Status: st.STATUS_IDLE,
	})
	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err != nil {
		log.Println(err)
		bot.sendGeneralError(chatId)
		return
	}

	oldDate := bot.getDateFromNullString(task.Date)
	midnight := bot.getMidnightFromDate(oldDate)
	newDate := sql.NullString{
		String: midnight.Format(DateWithTimeFormat),
		Valid:  true,
	}
	err = bot.taskService.UpdateTask(context.Background(), task, units.TaskPatch{
		Date: &newDate,
	})
	if err != nil {
		log.Println(err)
		bot.sendGeneralError(chatId)
		return
	}

	message := getEditingTaskInfo(task.Title, midnight)
	keyboard := buildEditTaskKeyboard(midnight, taskId)

	bot.editMessage(chatId, messageId, message, keyboard)
}

func (bot *Bot) completeTask(chatId int64, messageId int, taskId int) {
	_, err := bot.taskService.CompleteTask(context.Background(), taskId)
	if err == nil {
		message, keyboard := bot.getTasksListWithHeader()

		bot.editMessage(chatId, messageId, message, keyboard)
	} else {
		log.Println(err)
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) removeAllDoneTasks(chatId int64, messageId int) {
	keyboard := bot.createRemoveAllDoneTasksConfirmationKeyboard()
	bot.editMessage(chatId, messageId, TextRemoveAllDoneTasksConfirmation, keyboard)
}

func (bot *Bot) deleteTask(chatId int64, messageId int, taskId int) {
	err := bot.taskService.RemoveByID(context.Background(), taskId)
	if err == nil {
		bot.showTaskListInSameMessage(chatId, messageId)
	} else {
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) showTaskListInSameMessage(chatId int64, messageId int) {
	message, keyboard := bot.getTasksListWithHeader()

	bot.editMessage(chatId, messageId, message, keyboard)
}

func (bot *Bot) removeAllDoneTasksYes(chatId int64, messageId int) {
	err := bot.taskService.RemoveCompete(context.Background())
	if err == nil {
		message, keyboard := bot.getTasksListWithHeader()

		bot.editMessage(chatId, messageId, message, keyboard)
	} else {
		log.Println(err)
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) editTask(chatId int64, messageId int, taskId int) {
	task, err := bot.taskService.TaskByID(context.Background(), uint(taskId))
	if err == nil {
		date := bot.getDateFromNullString(task.Date)
		message := getEditingTaskInfo(task.Title, date)
		keyboard := buildEditTaskKeyboard(date, taskId)

		bot.editMessage(chatId, messageId, message, keyboard)
	} else {
		log.Println(err)
		bot.sendGeneralError(chatId)
	}
}

func (bot *Bot) createRemoveAllDoneTasksConfirmationKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return bot.createYesNoKeyboard(CQTaskRemoveAllDoneYes, CQTaskRemoveAllDoneNo)
}

// common handlers
func (bot *Bot) sendGeneralError(chatId int64) {
	bot.sendMessage(chatId, TextGeneralError, nil)
}

func (bot *Bot) sendParseError(chatId int64) {
	bot.sendMessage(chatId, TextParseError, nil)
}
