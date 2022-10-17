package bot

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maxwww/family_bot/units"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DateWithTimeFormat   = "2006-01-02 15:04"
	TimeFormat           = "15:04"
	DateWithTimeFormatUA = "02.01.2006 15:04"
	DateFormatUA         = "02.01.2006"

	DeFaultNotification = OneHourNotification
)

var DayNames = map[time.Weekday]string{
	time.Monday:    "Понеділок",
	time.Tuesday:   "Вівторок",
	time.Wednesday: "Середа",
	time.Thursday:  "Четвер",
	time.Friday:    "П'ятниця",
	time.Saturday:  "Субота",
	time.Sunday:    "Неділя",
}

var daysMap = map[time.Weekday][]string{
	time.Monday:    {"понеділок", "пн", "пон", "monday", "mon"},
	time.Tuesday:   {"вівторок", "вт", "вів", "tuesday", "tue", "tues"},
	time.Wednesday: {"середа", "ср", "сер", "wednesday", "wed"},
	time.Thursday:  {"четвер", "чт", "чет", "thursday", "thu", "thur", "thurs"},
	time.Friday:    {"п'ятниця", "пт", "п’ятниця", "пʼятниця", "пятниця", "п'ят", "п’ят", "пʼят", "пят", "friday", "fri"},
	time.Saturday:  {"субота", "сб", "суб", "saturday", "sat"},
	time.Sunday:    {"неділя", "нд", "нед", "sunday", "sun"},
}

var reMap map[time.Weekday][]*regexp.Regexp
var spaceRe *regexp.Regexp
var tomorrowRe *regexp.Regexp
var todayRe *regexp.Regexp
var dateRe *regexp.Regexp
var timeRe *regexp.Regexp

func init() {
	spaceRe = regexp.MustCompile(`\s+`)
	tomorrowRe = regexp.MustCompile(`(?i)(\s|^)завтра(\s|$)`)
	todayRe = regexp.MustCompile(`(?i)(\s|^)сьогодні(\s|$)`)
	dateRe = regexp.MustCompile(`([0123]?[0-9])(\/|\.)([01]?[0-9])((\/|\.)([0-9]{4}|[0-9]{2}))?`)
	timeRe = regexp.MustCompile(`([012]?[0-9]):([012345]?[0-9])`)
	reMap = make(map[time.Weekday][]*regexp.Regexp)
	for day, value := range daysMap {
		reMap[day] = []*regexp.Regexp{}
		for _, word := range value {
			r, err := regexp.Compile(`(?i)(\s|^)` + word + `(\s|$)`)
			if err != nil {
				panic(err)
			}
			reMap[day] = append(reMap[day], r)
		}
	}
}

func (bot *Bot) deleteMessage(chatId int64, messageId int) {
	msg := tgbotapi.NewDeleteMessage(chatId, messageId)
	_, err := bot.BotAPI.Request(msg)
	if err != nil {
		log.Println(err)
	}
}

func (bot *Bot) sendMessage(chatId int64, message string, keyboard *tgbotapi.InlineKeyboardMarkup, parseMode string) {
	msg := tgbotapi.NewMessage(chatId, message)
	if parseMode == "" {
		parseMode = "html"
	}
	msg.ParseMode = parseMode
	if keyboard != nil && len((*keyboard).InlineKeyboard) > 0 {
		msg.ReplyMarkup = keyboard
	}
	_, err := bot.BotAPI.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (bot *Bot) editMessage(chatId int64, messageId int, message string, keyboard *tgbotapi.InlineKeyboardMarkup, parseMode string) {
	msg := tgbotapi.NewEditMessageText(chatId, messageId, message)
	if parseMode == "" {
		parseMode = "html"
	}
	msg.ParseMode = parseMode
	if keyboard != nil && len((*keyboard).InlineKeyboard) > 0 {
		msg.ReplyMarkup = keyboard
	}
	_, err := bot.BotAPI.Request(msg)
	if err != nil {
		log.Println(err)
	}
}

func (bot *Bot) getMidnightFromDate(date *time.Time) *time.Time {
	var newDate *time.Time
	if date != nil {
		tmp := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, bot.loc)
		newDate = &tmp
	}

	return newDate
}

func (bot *Bot) getTasksListWithHeader() (string, *tgbotapi.InlineKeyboardMarkup) {
	message := TextTasksListHeader
	list, keyboard := bot.buildTasksList()
	if list != "" {
		message += "\n\n"
		message += list
	} else {
		message = TextTasksListEmpty
	}
	message += "\n"
	message += bot.buildToday()

	return message, keyboard
}

func (bot *Bot) buildTasksList() (string, *tgbotapi.InlineKeyboardMarkup) {
	list := ""
	var keyboard tgbotapi.InlineKeyboardMarkup
	tasks, _ := bot.taskService.Tasks(context.Background(), units.TaskFilter{})
	if len(tasks) > 0 {
		var tasksButtons [][]tgbotapi.InlineKeyboardButton
		var row []tgbotapi.InlineKeyboardButton
		hasDone := false
		now := time.Now().In(bot.loc)
		message := ""

		sort.Slice(tasks, func(i, j int) bool {
			if !tasks[i].Date.Valid && !tasks[j].Date.Valid {
				return tasks[i].ID < tasks[j].ID
			}

			if !tasks[i].Date.Valid {
				return false
			}

			if !tasks[j].Date.Valid {
				return true
			}

			iDate, _ := time.ParseInLocation(DateWithTimeFormat, tasks[i].Date.String[:10]+" "+tasks[i].Date.String[11:16], bot.loc)
			jDate, _ := time.ParseInLocation(DateWithTimeFormat, tasks[j].Date.String[:10]+" "+tasks[j].Date.String[11:16], bot.loc)
			isIDateWithTime := iDate.Hour() != 0 || iDate.Minute() != 0
			isJDateWithTime := jDate.Hour() != 0 || jDate.Minute() != 0

			if isIDateWithTime && !isJDateWithTime {
				if iDate.Year() == jDate.Year() && iDate.Month() == jDate.Month() && iDate.Day() == jDate.Day() {
					return true
				}
				return iDate.Before(jDate)
			}
			if !isIDateWithTime && isJDateWithTime {
				if iDate.Year() == jDate.Year() && iDate.Month() == jDate.Month() && iDate.Day() == jDate.Day() {
					return false
				}
				return iDate.Before(jDate)
			}

			return iDate.Before(jDate)
		})

		for i, v := range tasks {
			checkBox := TextCheckbox
			if v.Done {
				checkBox = TextComplete
				hasDone = true
				message += "<s>"
			}

			text := fmt.Sprintf("%d %s", i+1, checkBox)
			action := fmt.Sprintf("%s:%d", CQTaskComplete, v.ID)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(text, action))
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf(TextSettings), fmt.Sprintf(CQTaskEdit+":%d", v.ID)))
			if len(row) == 6 {
				tasksButtons = append(tasksButtons, row)
				row = []tgbotapi.InlineKeyboardButton{}
			}
			message += fmt.Sprintf("%d. %s", i+1, v.Title)
			if v.Date.Valid {
				date, _ := time.ParseInLocation(DateWithTimeFormat, v.Date.String[:10]+" "+v.Date.String[11:16], bot.loc)

				format := DateWithTimeFormatUA
				timeFormat := TimeFormat
				if date.Hour() == 0 && date.Minute() == 0 {
					format = DateFormatUA
					timeFormat = ""
				}
				if date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day() {
					message += fmt.Sprintf(" (" + TextToday)
					if timeFormat != "" {
						message += fmt.Sprintf(" %s", date.Format(timeFormat))
					}
					message += ")"
				} else {
					tomorrow := now.AddDate(0, 0, 1)
					if date.Year() == tomorrow.Year() && date.Month() == tomorrow.Month() && date.Day() == tomorrow.Day() {
						message += fmt.Sprintf(" (" + TextTomorrow)
						if timeFormat != "" {
							message += fmt.Sprintf(" %s", date.Format(timeFormat))
						}
						message += ")"
					} else {
						message += fmt.Sprintf(" (%s)", date.Format(format))
					}
				}

			}
			if v.Done {
				message += "</s>"
			}
			message += "\n"
		}
		if len(row) > 0 {
			tasksButtons = append(tasksButtons, row)
		}
		if hasDone {
			tasksButtons = append(tasksButtons, []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(TextActionRemoveAllDoneTasks, CQTaskRemoveAllDone),
			})
		}

		list += message
		keyboard = tgbotapi.NewInlineKeyboardMarkup(tasksButtons...)
	}

	return list, &keyboard
}

func (bot *Bot) buildToday() string {
	now := time.Now().In(bot.loc)
	return fmt.Sprintf(TextTodayDate, strings.ToLower(DayNames[now.Weekday()]), now.Format(DateFormatUA))
}

func (bot *Bot) createYesNoKeyboard(yesAction, noAction string) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData(TextActionYes, yesAction), tgbotapi.NewInlineKeyboardButtonData(TextActionNo, noAction)},
		},
	}
}

func (bot *Bot) getDateFromNullString(str sql.NullString) *time.Time {
	var date *time.Time
	if str.Valid && str.String != "" {
		date = bot.getDateFromString(str.String)
	}

	return date
}

func (bot *Bot) getDateFromString(str string) *time.Time {
	t, _ := time.ParseInLocation(DateWithTimeFormat, str[:10]+" "+str[11:16], bot.loc)

	return &t
}

func (bot *Bot) findDate(input string) (*time.Time, string, bool, error) {
	input = trim(input)
	var date *time.Time
	isDateFound := false

	var weekDay time.Weekday = -1
	for day, s := range reMap {
		for _, r := range s {
			index := r.FindStringIndex(input)
			if len(index) > 0 {
				input = input[0:index[0]] + " " + input[index[1]:]
				weekDay = day
			}
		}
	}

	now := time.Now().In(bot.loc)
	today := bot.getMidnightFromDate(&now)
	parsedDate := bot.getMidnightFromDate(&now)

	if weekDay != -1 {
		todayWeekDay := today.Weekday()
		delta := 0
		if weekDay <= todayWeekDay {
			delta = 7 - int(todayWeekDay) + int(weekDay)
		} else {
			delta = int(weekDay) - int(todayWeekDay)
		}
		*parsedDate = parsedDate.AddDate(0, 0, delta)

		date = parsedDate
	} else {
		tomorrowIndex := tomorrowRe.FindStringIndex(input)
		if len(tomorrowIndex) > 0 {
			input = input[0:tomorrowIndex[0]] + " " + input[tomorrowIndex[1]:]
			*parsedDate = parsedDate.AddDate(0, 0, 1)
			date = parsedDate
		} else {
			todayIndex := todayRe.FindStringIndex(input)
			if len(todayIndex) > 0 {
				input = input[0:todayIndex[0]] + " " + input[todayIndex[1]:]
				date = parsedDate
			} else {
				match := dateRe.FindStringSubmatch(input)
				if len(match) > 0 {
					input = dateRe.ReplaceAllString(input, "")
					year := today.Year()
					if match[6] != "" {
						if len(match[6]) == 2 {
							match[6] = "20" + match[6]
						}
						year_, err := strconv.Atoi(match[6])
						if err != nil {
							return nil, "", isDateFound, units.DataParsingError
						}
						year = year_
					}

					month, err := strconv.Atoi(match[3])
					if err != nil {
						return nil, "", isDateFound, units.DataParsingError
					}

					day, err := strconv.Atoi(match[1])
					if err != nil {
						return nil, "", isDateFound, units.DataParsingError
					}

					tmpParsedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, bot.loc)
					if tmpParsedDate.Before(*today) {
						if match[6] != "" {
							return nil, "", isDateFound, units.DataParsingError
						}
						tmpParsedDate = tmpParsedDate.AddDate(1, 0, 0)
					}
					date = &tmpParsedDate
				}
			}
		}
	}

	if date == nil {
		date = today
	} else {
		isDateFound = true
	}

	timeMatch := timeRe.FindStringSubmatch(input)
	if len(timeMatch) > 0 {
		input = timeRe.ReplaceAllString(input, "")

		hour, err := strconv.Atoi(timeMatch[1])
		if err != nil {
			return nil, "", isDateFound, units.DataParsingError
		}

		minutes, err := strconv.Atoi(timeMatch[2])
		if err != nil {
			return nil, "", isDateFound, units.DataParsingError
		}

		*date = date.Add(time.Duration(hour) * time.Hour)
		*date = date.Add(time.Duration(minutes) * time.Minute)
		if date.Before(now) {
			*date = date.AddDate(0, 0, 1)
		}
	} else if !isDateFound {
		date = nil
	}

	return date, input, isDateFound, nil
}
