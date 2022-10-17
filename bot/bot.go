package bot

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/maxwww/family_bot/postgres"
	st "github.com/maxwww/family_bot/state"
	"github.com/maxwww/family_bot/units"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	BotAPI       *tgbotapi.BotAPI
	loc          *time.Location
	subscribers  []int64
	userService  units.UserService
	taskService  units.TaskService
	stateService st.StateServiceI
}

func NewBot(botAPI *tgbotapi.BotAPI, db *postgres.DB, subscribers []int64, loc *time.Location) *Bot {
	bot := Bot{
		BotAPI:      botAPI,
		loc:         loc,
		subscribers: subscribers,
	}

	bot.userService = postgres.NewUserService(db)
	bot.taskService = postgres.NewTaskService(db)
	bot.stateService = st.NewStateService()

	return &bot
}

func (bot *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.BotAPI.GetUpdatesChan(u)

	err := bot.RegisterCrons()
	if err != nil {
		return err
	}

	for update := range updates {
		go bot.handleUpdate(update)
	}

	return nil
}

func (bot *Bot) handleUpdate(update tgbotapi.Update) {

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	var fromUser *tgbotapi.User
	var chatId int64

	if update.CallbackQuery != nil {
		fromUser = update.CallbackQuery.From
		chatId = update.CallbackQuery.Message.Chat.ID
	} else {
		fromUser = update.Message.From
		chatId = update.Message.Chat.ID
	}

	user, err := bot.userService.UserByTelegramID(context.Background(), uint(fromUser.ID))

	if err != nil {
		if err != units.ErrNotFound {
			// TODO: handle error
			log.Println(err)
			bot.sendGeneralError(chatId)
			return
		}
		err = bot.userService.CreateUser(context.Background(), &units.User{
			TelegramID: uint(update.Message.From.ID),
			FirstName:  update.Message.From.FirstName,
			LastName:   update.Message.From.LastName,
			UserName:   update.Message.From.UserName,
		})
		if err != nil {
			bot.sendGeneralError(chatId)
			return
		}

		user, err = bot.userService.UserByTelegramID(context.Background(), uint(fromUser.ID))
	}

	isSubscriber := false
	for _, v := range bot.subscribers {
		if v == int64(user.TelegramID) {
			isSubscriber = true
		}
	}
	if !isSubscriber {
		return
	}

	state := bot.stateService.GetUserState(int(user.TelegramID))

	if update.CallbackQuery != nil {
		msg := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		_, err := bot.BotAPI.Request(msg)
		if err != nil {
			log.Println(err)
		}

		command := update.CallbackQuery.Data
		id := 0
		commandWithParam := strings.Split(update.CallbackQuery.Data, ":")
		param := 0
		if len(commandWithParam) > 1 {
			command = commandWithParam[0]
			id, _ = strconv.Atoi(commandWithParam[1])
			if len(commandWithParam) > 2 {
				param, _ = strconv.Atoi(commandWithParam[2])
			}
		}

		switch command {
		case CQNewTaskSave:
			bot.saveNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskEditTitle:
			bot.editTitleNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskEditDay:
			bot.editDayNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskRemoveDay:
			bot.removeDayNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskEditTime:
			bot.editTimeNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskRemoveTime:
			bot.removeTimeNewTask(state, chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskCancel:
			bot.cancelNewTask(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID))
		case CQNewTaskSetNotifications:
			bot.setNotificationsNewTask(state, chatId, update.CallbackQuery.Message.MessageID, param, int(user.TelegramID))
		case CQTaskComplete:
			bot.completeTask(chatId, update.CallbackQuery.Message.MessageID, id)
		case CQTaskRemoveAllDone:
			bot.removeAllDoneTasks(chatId, update.CallbackQuery.Message.MessageID)
		case CQTaskRemoveAllDoneNo:
			bot.showTaskListInSameMessage(chatId, update.CallbackQuery.Message.MessageID)
		case CQTaskRemoveAllDoneYes:
			bot.removeAllDoneTasksYes(chatId, update.CallbackQuery.Message.MessageID)
		case CQTaskEdit:
			bot.editTask(chatId, update.CallbackQuery.Message.MessageID, id)
		case CQTaskEditOk:
			bot.showTaskListInSameMessage(chatId, update.CallbackQuery.Message.MessageID)
		case CQTaskEditEditTitle:
			bot.editTaskTitle(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id)
		case CQTaskEditEditDay:
			bot.editTaskDay(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id)
		case CQTaskEditEditTime:
			bot.editTaskTime(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id)
		case CQTaskEditRemoveDay:
			bot.removeTaskDay(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id)
		case CQTaskEditRemoveTime:
			bot.removeTaskTime(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id)
		case CQTaskEditDeleteTask:
			bot.deleteTask(chatId, update.CallbackQuery.Message.MessageID, id)
		case CQTaskEditSetNotifications:
			bot.setNotifications(chatId, update.CallbackQuery.Message.MessageID, int(user.TelegramID), id, param)
		default:
			bot.sendGeneralError(chatId)
		}
	} else if update.Message.IsCommand() {
		switch update.Message.Command() {
		case commandStart, commandHelp:
			bot.handleStartCommand(chatId)
		case commandList:
			bot.handleListCommand(chatId)
		case commandCancel:
			bot.handleCancelCommand(chatId, int(user.TelegramID))
		case commandSubscribe:
			bot.handleSubscribeCommand(chatId, true, user)
		case commandUnsubscribe:
			bot.handleSubscribeCommand(chatId, false, user)
		default:
			bot.handleUnknownCommand(chatId)
		}
	} else {
		switch state.Status {
		case st.STATUS_IDLE:
			bot.handleIdleMessage(chatId, user, update.Message.Text)
		case st.STATUS_ADD_TASK_WAIT_TITLE:
			bot.handleNewTaskEditTitle(chatId, user, state.Task, update.Message.Text)
		case st.STATUS_ADD_TASK_WAIT_DATE:
			bot.handleNewTaskEditDate(chatId, user, state.Task, update.Message.Text)
		case st.STATUS_ADD_TASK_WAIT_TIME:
			bot.handleNewTaskEditTime(chatId, user, state.Task, update.Message.Text)
		case st.STATUS_EDIT_TASK_WAIT_TITLE:
			bot.handleEditTaskEditTitle(chatId, int(user.TelegramID), update.Message.Text, state.Task.ID)
		case st.STATUS_EDIT_TASK_WAIT_DATE:
			bot.handleEditTaskEditDate(chatId, int(user.TelegramID), update.Message.Text, state.Task.ID)
		case st.STATUS_EDIT_TASK_WAIT_TIME:
			bot.handleEditTaskEditTime(chatId, int(user.TelegramID), update.Message.Text, state.Task.ID)
		}
	}
}
