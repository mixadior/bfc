package botform

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func RenderChoiceControl(bckc BotChoiceKeyboardControl) tgbotapi.Chattable {
	m := tgbotapi.NewMessage(0, "")
	m.Text = bckc.Title

	row := []tgbotapi.KeyboardButton{}
	rows := [][]tgbotapi.KeyboardButton{}

	controlRow := []tgbotapi.KeyboardButton{}
	ssmChoices := SortedStringMap(bckc.ControlChoices)
	ssmChoices.Range(func(k, choice string) bool {
		controlRow = append(controlRow, tgbotapi.NewKeyboardButton(choice))
		return true
	}, false)

	if len(controlRow) > 0 {
		rows = append(rows, controlRow)
	}
	ssmMainChoices := SortedStringMap(bckc.Choices)
	ssmMainChoices.Range(func(k, choice string) bool {
		row = append(row, tgbotapi.NewKeyboardButton(choice))
		if len(row) >= bckc.ButtonsPerRow {
			rows = append(rows, row)
			row = []tgbotapi.KeyboardButton{}
		}
		return true
	}, true)

	if len(row) > 0 {
		rows = append(rows, row)
	}

	rk := tgbotapi.NewReplyKeyboard(rows...)
	m.ReplyMarkup = rk
	return m
}

func RenderTextControl(text BotTextKeyboardControl) tgbotapi.Chattable {
	m := tgbotapi.NewMessage(0, "")
	m.Text = text.Title

	row := []tgbotapi.KeyboardButton{}
	rows := [][]tgbotapi.KeyboardButton{}

	ssmChoices := SortedStringMap(text.ControlChoices)
	ssmChoices.Range(func(k, choice string) bool {
		row = append(row, tgbotapi.NewKeyboardButton(choice))
		return true
	}, false)

	rows = append(rows, row)

	rk := tgbotapi.NewReplyKeyboard(rows...)
	m.ReplyMarkup = rk
	return m
}

func RenderMultiChoiceControl(multichoice BotMultiChoiceKeyboardControl) tgbotapi.Chattable {
	bckc := multichoice
	m := tgbotapi.NewMessage(0, "")
	m.Text = bckc.Title

	row := []tgbotapi.KeyboardButton{}
	rows := [][]tgbotapi.KeyboardButton{}

	controlRow := []tgbotapi.KeyboardButton{}
	ssmChoices := SortedStringMap(bckc.ControlChoices)
	ssmChoices.Range(func(k, choice string) bool {
		controlRow = append(controlRow, tgbotapi.NewKeyboardButton(choice))
		return true
	}, false)

	if len(controlRow) > 0 {
		rows = append(rows, controlRow)
	}
	ssmMainChoices := SortedStringMap(bckc.Choices)
	ssmMainChoices.Range(func(k, choice string) bool {
		_, existsInSelected := bckc.SelectedValues[choice]
		if existsInSelected {
			choice = bckc.SelectedPrefix + choice
		}
		row = append(row, tgbotapi.NewKeyboardButton(choice))
		if len(row) >= bckc.ButtonsPerRow {
			rows = append(rows, row)
			row = []tgbotapi.KeyboardButton{}
		}
		return true
	}, true)

	if len(row) > 0 {
		rows = append(rows, row)
	}

	rk := tgbotapi.NewReplyKeyboard(rows...)
	m.ReplyMarkup = rk
	return m
}

func RenderToTgBotApi(control BotControl) tgbotapi.Chattable {
	switch control.(type) {
	case BotChoiceKeyboardControl:
		return RenderChoiceControl(control.(BotChoiceKeyboardControl))
	case BotTextKeyboardControl:
		return RenderTextControl(control.(BotTextKeyboardControl))
	case BotMultiChoiceKeyboardControl:
		return RenderMultiChoiceControl(control.(BotMultiChoiceKeyboardControl))
	default:
		panic("Unsupported control type to render")
	}
}
