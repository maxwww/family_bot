package bot

const (
	TextActionOk                 = "✅ OK"
	TextActionYes                = "✅ Так"
	TextActionNo                 = "❌ Ні"
	TextActionEditTitle          = "✏ назву"
	TextActionCancel             = "❌ Відмінити"
	TextActionDelete             = "❌ Видалити"
	TextActionEditDay            = "✏ дату"
	TextActionRemoveDay          = "❌ дату"
	TextActionEditTime           = "✏ час"
	TextActionRemoveTime         = "❌ час"
	TextCheckbox                 = "☑"
	TextComplete                 = "✅"
	TextSettings                 = "⚙"
	TextActionRemoveAllDoneTasks = "❌ Видалити всі виконані справи"

	TextGeneralError                   = "Сталася помилка. Спробуйте пізніше."
	TextParseError                     = "Вибач, але я не розумію."
	TextNewTaskAdded                   = "Додано нову справу:"
	TextTaskDescription                = "Назва: %s\nДень: %s\nЧас: %s"
	TextSendNewTitle                   = "Стара назва - `%s`\nНапишіть мені нову назву справи"
	TextSendNewDay                     = "Напишіть мені нову дату справи"
	TextSendNewTime                    = "Напишіть мені новий час справи"
	TextTasksListHeader                = "Ось список усіх справ:"
	TextTasksListEmpty                 = "Задачі відсутні"
	TextToday                          = "сьогодні"
	TextTomorrow                       = "завтра"
	TextTodayDate                      = "До речі сьогодні %s %s"
	TextRemoveAllDoneTasksConfirmation = "Видалити всі виконані справи?"
	TextTaskEditing                    = "Редагування справи:"
	TextNewTaskEditing                 = "Нова справа. Перевірьте заповнені поля та натисніть OK:"
	TextUnknownCommand                 = "На жаль, я не знаю такої команди. Скористайтеся меню або довідкою - /help"
	TextInOneHour                      = "Справа \"%s\" через одну годину"
	TextInThirtyMinutes                = "Справа \"%s\" через 30 хв"
	TextInFiveMinutes                  = "Справа \"%s\" через 5 хв"
	TextInInstantly                    = "Справа \"%s\" розпочалася"
	TextCancel                         = "Охрана, отмєна"
	TextSubscriptionsOn                = "Сповіщення увімкнено.\nАби вимкунити сповіщення скористайся /unsubscribe командою."
	TextSubscriptionsOff               = "Сповіщення вимкнуто.\nАби увімкнути сповіщення скористайся /subscribe командою."
	TextNotificationOneHour            = "1 год"
	TextNotificationThirtyMinutes      = "30 хв"
	TextNotificationFiveMinutes        = "5 хв"
	TextNotificationInstantly          = "0 хв"
	TextStartMessage                   = `Я, 🤖. Я можу допомагати тобі слідкувати за сімейними справами.

Ось список моїх команд:
/list - переглянути список сімейних справ
/cancel - відмінити поточну операцію

Залишились питання чи є пропозиція? Звертайся до цього контакту - @msfilo`
)
