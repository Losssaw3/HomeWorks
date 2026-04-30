package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func getInlineButtons() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ввести город", "city"),
			tgbotapi.NewInlineKeyboardButtonData("Ввести координаты", "coords"),
			tgbotapi.NewInlineKeyboardButtonData("Назад", "return"),
		),
	)
}

func getOnlyReturnButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "return"),
		),
	)
}
