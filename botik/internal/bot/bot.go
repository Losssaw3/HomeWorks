package bot

import (
	"log"
	"strings"

	"bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	greeting         = "Выберите способ получения погоды (по названию города или по координатам) , для возврата нажмите назад"
	errorMessage     = "Что-то пошло не так проверьте введенные данные и повторите попытку"
	errorButtonUsage = "Пожалуйста завершите текущее действие или нажмите назад"
	tipForCoods      = "Введите координаты: 55.72, 37.61"
	tipForCity       = "Введите город и код страны: Moscow, RU"
	startState       = iota
	stateWaitForCoords
	stateStateWaitForCity
)

type weatherServiceI interface {
	GetWeatherByCoords(lat, long string) (*services.WeatherResponce, error)
	GetWeatherByCity(cityName, countryCode string) (*services.WeatherResponce, error)
}

type State int

var userStates = make(map[int64]State)

type tgBot struct {
	bot     *tgbotapi.BotAPI
	service weatherServiceI
}

func NewBot(token string, service weatherServiceI) *tgBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	return &tgBot{bot: bot, service: service}
}

func (b *tgBot) StartBot() {

	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		var userId int64
		if update.Message != nil {
			userId = update.Message.Chat.ID
		} else if update.CallbackQuery != nil {
			userId = update.CallbackQuery.Message.Chat.ID
		} else {
			continue
		}
		if _, ok := userStates[userId]; !ok {
			userStates[userId] = startState
		}

		if update.Message != nil {
			b.handleMessage(update.Message)
		}

		if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}
}

func (b *tgBot) handleCallback(cb *tgbotapi.CallbackQuery) {
	if cb.Message == nil {
		return
	}
	callbackConfig := tgbotapi.NewCallback(cb.ID, "")
	b.bot.Request(callbackConfig)

	userId := cb.Message.Chat.ID
	data := cb.Data

	switch data {
	case "city":
		if state := userStates[userId]; state != startState {
			msg := tgbotapi.NewMessage(userId, errorButtonUsage)
			newButtons := getOnlyReturnButton()
			msg.ReplyMarkup = &newButtons
			b.bot.Send(msg)
		} else {
			userStates[userId] = stateStateWaitForCity
			edit := tgbotapi.NewMessage(userId, tipForCity)
			newButtons := getOnlyReturnButton()
			edit.ReplyMarkup = &newButtons
			b.bot.Send(edit)
		}

	case "coords":
		if state := userStates[userId]; state != startState {
			msg := tgbotapi.NewMessage(userId, errorButtonUsage)
			newButtons := getOnlyReturnButton()
			msg.ReplyMarkup = &newButtons
			b.bot.Send(msg)
		} else {
			userStates[userId] = stateWaitForCoords
			edit := tgbotapi.NewMessage(userId, tipForCoods)
			newButtons := getOnlyReturnButton()
			edit.ReplyMarkup = &newButtons
			b.bot.Send(edit)
		}

	case "return":
		userStates[userId] = startState
		msg := tgbotapi.NewMessage(userId, greeting)
		msg.ReplyMarkup = getInlineButtons()
		b.bot.Send(msg)
	}
}

func (b *tgBot) handleMessage(cb *tgbotapi.Message) {
	text := cb.Text
	userId := cb.Chat.ID

	switch userStates[userId] {
	case startState:
		msg := tgbotapi.NewMessage(userId, greeting)
		msg.ReplyMarkup = getInlineButtons()
		b.bot.Send(msg)

	case stateStateWaitForCity:
		arr := strings.Split(text, ",")
		if len(arr) != 2 {
			msg := tgbotapi.NewMessage(userId, errorMessage)
			msg.ReplyMarkup = getOnlyReturnButton()
			b.bot.Send(msg)
		} else {
			result, err := b.service.GetWeatherByCity(arr[0], arr[1])
			if err != nil {
				msg := tgbotapi.NewMessage(userId, errorMessage)
				msg.ReplyMarkup = getOnlyReturnButton()
				b.bot.Send(msg)
				return
			}
			msg := tgbotapi.NewMessage(userId, result.Prefix+result.Report)
			b.bot.Send(msg)
			userStates[userId] = startState

			Nmsg := tgbotapi.NewMessage(userId, greeting)
			Nmsg.ReplyMarkup = getInlineButtons()
			b.bot.Send(Nmsg)
		}
	case stateWaitForCoords:
		arr := strings.Split(text, ",")
		if len(arr) != 2 {
			msg := tgbotapi.NewMessage(userId, errorMessage)
			msg.ReplyMarkup = getOnlyReturnButton()
			b.bot.Send(msg)
		} else {
			result, err := b.service.GetWeatherByCoords(arr[0], arr[1])
			if err != nil {
				msg := tgbotapi.NewMessage(userId, errorMessage)
				msg.ReplyMarkup = getOnlyReturnButton()
				b.bot.Send(msg)
				return
			}
			msg := tgbotapi.NewMessage(userId, result.Prefix+result.Report)
			b.bot.Send(msg)
			userStates[userId] = startState

			Nmsg := tgbotapi.NewMessage(userId, greeting)
			Nmsg.ReplyMarkup = getInlineButtons()
			b.bot.Send(Nmsg)

		}
	}

}
