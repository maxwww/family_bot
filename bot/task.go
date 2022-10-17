package bot

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	st "github.com/maxwww/family_bot/state"
	"github.com/maxwww/family_bot/units"
	"strings"
	"time"
)

const (
	OneHourNotification = 1 << iota
	ThirtyMinutesNotification
	FiveMinutesNotification
	InstantlyNotification
)

func createTaskStateWithDate(task *st.Task) *units.Task {
	newTask := units.Task{
		Title:         task.Title,
		Date:          sql.NullString{},
		Done:          false,
		Notifications: task.Notifications,
	}

	if task.Date != nil {
		newTask.Date = sql.NullString{
			String: task.Date.Format(DateWithTimeFormat),
			Valid:  true,
		}
	}

	return &newTask
}

func getSavedTaskInfo(title string, date *time.Time) string {
	return getOneTaskInfo(TextNewTaskAdded, title, date)
}

func getEditingTaskInfo(title string, date *time.Time) string {
	return getOneTaskInfo(TextTaskEditing, title, date)
}

func getNewTaskInfo(title string, date *time.Time) string {
	return getOneTaskInfo(TextNewTaskEditing, title, date)
}

func getOneTaskInfo(header string, title string, date *time.Time) string {
	dayString, timeString := getDayAndTime(date)
	return fmt.Sprintf("%s\n\n%s", header, getTaskDescription(title, dayString, timeString))
}

func getDayAndTime(date *time.Time) (string, string) {
	dayString := "-"
	timeString := "-"
	if date != nil {
		dayString = date.Format(DateFormatUA)
		if date.Hour() != 0 || date.Minute() != 0 {
			timeString = date.Format(TimeFormat)
		}
	}

	return dayString, timeString
}

func getTaskDescription(title, dayString, timeString string) string {
	return fmt.Sprintf(TextTaskDescription, title, dayString, timeString)
}

func buildEditTaskKeyboard(date *time.Time, notifications int, taskId int) *tgbotapi.InlineKeyboardMarkup {
	editDayData := fmt.Sprintf(CQTaskEditEditDay+":%d", taskId)
	removeDayData := fmt.Sprintf(CQTaskEditRemoveDay+":%d", taskId)
	editTimeData := fmt.Sprintf(CQTaskEditEditTime+":%d", taskId)
	removeTimeData := fmt.Sprintf(CQTaskEditRemoveTime+":%d", taskId)
	OKData := CQTaskEditOk
	editTitleData := fmt.Sprintf(CQTaskEditEditTitle+":%d", taskId)
	cancelDeleteData := fmt.Sprintf(CQTaskEditDeleteTask+":%d", taskId)
	setNotifications := fmt.Sprintf(CQTaskEditSetNotifications+":%d", taskId)
	cancelDeleteAction := TextActionDelete

	if taskId == 0 {
		editDayData = CQNewTaskEditDay
		removeDayData = CQNewTaskRemoveDay
		editTimeData = CQNewTaskEditTime
		removeTimeData = CQNewTaskRemoveTime
		OKData = CQNewTaskSave
		editTitleData = CQNewTaskEditTitle
		cancelDeleteData = CQNewTaskCancel
		cancelDeleteAction = TextActionCancel
		setNotifications = CQNewTaskSetNotifications + ":"
	}

	dateButtons := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(TextActionEditDay, editDayData)}
	if date != nil {
		dateButtons = append(dateButtons, tgbotapi.NewInlineKeyboardButtonData(TextActionRemoveDay, removeDayData))
		dateButtons = append(dateButtons, tgbotapi.NewInlineKeyboardButtonData(TextActionEditTime, editTimeData))
		if date.Hour() != 0 || date.Minute() != 0 {
			dateButtons = append(dateButtons, tgbotapi.NewInlineKeyboardButtonData(TextActionRemoveTime, removeTimeData))
		}
	}
	var notificationsButtons []tgbotapi.InlineKeyboardButton

	for _, v := range []int{OneHourNotification, ThirtyMinutesNotification, FiveMinutesNotification, InstantlyNotification} {
		checkBox := TextCheckbox
		if (notifications & v) != 0 {
			checkBox = TextComplete
		}
		notificationsButtons = append(notificationsButtons, tgbotapi.NewInlineKeyboardButtonData(checkBox+" "+getNotificationLabel(v), fmt.Sprintf(setNotifications+":%d", v)))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(TextActionOk, OKData),
			tgbotapi.NewInlineKeyboardButtonData(TextActionEditTitle, editTitleData),
		),
		tgbotapi.NewInlineKeyboardRow(
			dateButtons...,
		),
		tgbotapi.NewInlineKeyboardRow(
			notificationsButtons...,
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(cancelDeleteAction, cancelDeleteData),
		),
	)

	return &keyboard
}

func trim(input string) (out string) {
	out = strings.Trim(input, " ")
	out = spaceRe.ReplaceAllString(out, " ")

	return
}

func getNotificationLabel(notification int) string {
	switch notification {
	case OneHourNotification:
		return TextNotificationOneHour
	case ThirtyMinutesNotification:
		return TextNotificationThirtyMinutes
	case FiveMinutesNotification:
		return TextNotificationFiveMinutes
	case InstantlyNotification:
		return TextNotificationInstantly
	}

	return ""
}
