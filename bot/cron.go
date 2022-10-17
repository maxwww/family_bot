package bot

import (
	"context"
	"fmt"
	"github.com/maxwww/family_bot/units"
	"github.com/robfig/cron/v3"
	"time"
)

func (bot *Bot) RegisterCrons() error {
	c := cron.New(cron.WithLocation(bot.loc))

	_, err := c.AddFunc("30 8 * * *", func() {
		for _, userId := range bot.subscribers {
			user, err := bot.userService.UserByTelegramID(context.Background(), uint(userId))
			if err == nil && user.Notifications {
				message, keyboard := bot.getTasksListWithHeader()
				bot.sendMessage(userId, message, keyboard, "")
			}
		}
	})
	if err != nil {
		return err
	}

	_, err = c.AddFunc("* * * * *", func() {
		tasks, _ := bot.taskService.Tasks(context.Background(), units.TaskFilter{})
		var taskForNotifications []struct {
			task         *units.Task
			notification int
		}
		now := time.Now().In(bot.loc)

		for _, notification := range []int{OneHourNotification, ThirtyMinutesNotification, FiveMinutesNotification, InstantlyNotification} {
			minutes := getMinutesFromNotificationType(notification)
			notificationTime := now.Add(time.Duration(minutes) * time.Minute)
			for _, v := range tasks {
				if (v.Notifications&notification) != 0 && v.Date.Valid && v.Date.String[:10]+" "+v.Date.String[11:16] == notificationTime.Format(DateWithTimeFormat) &&
					(notificationTime.Hour() != 0 || notificationTime.Minute() != 0) {
					taskForNotifications = append(taskForNotifications, struct {
						task         *units.Task
						notification int
					}{task: v, notification: notification})
				}
			}
		}

		for _, item := range taskForNotifications {
			textFormat := getFormatFromNotificationType(item.notification)
			message := fmt.Sprintf(textFormat, item.task.Title)
			for _, userId := range bot.subscribers {
				user, err := bot.userService.UserByTelegramID(context.Background(), uint(userId))
				if err == nil && user.Notifications {
					bot.sendMessage(userId, message, nil, "")
				}
			}
		}
	})
	if err != nil {
		return err
	}

	c.Start()

	return nil
}

func getMinutesFromNotificationType(notificationType int) int {
	switch notificationType {
	case OneHourNotification:
		return 60
	case ThirtyMinutesNotification:
		return 30
	case FiveMinutesNotification:
		return 5
	}

	return 0
}

func getFormatFromNotificationType(notificationType int) string {
	switch notificationType {
	case OneHourNotification:
		return TextInOneHour
	case ThirtyMinutesNotification:
		return TextInThirtyMinutes
	case FiveMinutesNotification:
		return TextInFiveMinutes
	}

	return TextInInstantly
}
