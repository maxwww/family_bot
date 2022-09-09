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

	_, err := c.AddFunc("50 14 * * *", func() {
		//_, err := c.AddFunc("30 8 * * *", func() {
		for _, userId := range bot.subscribers {
			message, keyboard := bot.getTasksListWithHeader()
			bot.sendMessage(userId, message, keyboard)
		}
	})
	if err != nil {
		return err
	}

	_, err = c.AddFunc("* * * * *", func() {
		tasks, _ := bot.taskService.Tasks(context.Background(), units.TaskFilter{})
		var taskForNotifications []*units.Task
		now := time.Now().In(bot.loc)
		now = now.Add(1 * time.Hour)
		for _, v := range tasks {
			if v.Date.Valid && v.Date.String[:10]+" "+v.Date.String[11:16] == now.Format(DateWithTimeFormat) {
				taskForNotifications = append(taskForNotifications, v)
			}
		}
		for _, task := range taskForNotifications {
			message := fmt.Sprintf(TextInOneHour, task.Title)
			for _, userId := range bot.subscribers {
				bot.sendMessage(userId, message, nil)
			}
		}
	})
	if err != nil {
		return err
	}

	c.Start()

	return nil
}
