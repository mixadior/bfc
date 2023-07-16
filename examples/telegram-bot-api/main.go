package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	botform "github.com/mixadior/bfc"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	updateChannel := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   100,
		Timeout: 60,
	})

	if err != nil {
		log.Fatal(err)
	}

	sessionStore := sync.Map{}
	bfc := loadBfc()
	dummySession := func() *Session { return &Session{BFCState: botform.NewState(map[string]string{}, botform.BfcKey{"root"}), BFCStarted: false} }

	for update := range updateChannel {
		if update.Message == nil {
			continue
		}

		chatId := update.Message.Chat.ID
		sessionRaw, _ := sessionStore.LoadOrStore(chatId, dummySession())
		messageText := update.Message.Text
		session := sessionRaw.(*Session)

		if messageText == "/wizard" {
			session = dummySession()
			session.BFCStarted = true
		}

		if !session.BFCStarted {
			replyText := "To start bfc send /wizard"
			sendReply(bot, update.Message.Chat.ID, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, replyText), true)
			continue
		}

		validationMessage := ""
		state := session.BFCState
		v, e, nextAction := bfc.ValidateAndTransform(state, messageText)
		if e != nil {
			validationMessage = e.Error()
			sendReply(bot, update.Message.Chat.ID, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, validationMessage), false)
		} else {
			bfc.CommitToState(v, nextAction, state)
		}

		if state.IsFinished {
			isOk := state.Values["is_ok"]
			if isOk == "yes" {
				prettyValues := bfc.GetPrettyValues(state)
				texts := []string{}
				texts = append(texts, "_"+tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Answers saved.")+"_")
				for keyTitle, keyValue := range prettyValues {
					texts = append(texts, "*"+tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, keyTitle)+"*"+": "+tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, keyValue))
				}
				m := tgbotapi.NewMessage(update.Message.Chat.ID, strings.Join(texts, "\n"))
				m.ParseMode = tgbotapi.ModeMarkdownV2
				m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				session.BFCStarted = false
				bot.Send(m)
			} else {
				m := tgbotapi.NewMessage(update.Message.Chat.ID, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Answers not saved."))
				m.ParseMode = tgbotapi.ModeMarkdownV2
				m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				bot.Send(m)
			}
		} else if state.IsCanceled {
			session.BFCStarted = false
			m := tgbotapi.NewMessage(update.Message.Chat.ID, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Canceled ((("))
			m.ParseMode = tgbotapi.ModeMarkdownV2
			m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			bot.Send(m)
		} else {
			controlContrainer, _ := bfc.Build(state)
			m := botform.RenderToTgBotApi(controlContrainer)

			mm := m.(tgbotapi.MessageConfig)
			mm.ChatID = update.Message.Chat.ID
			mm.ParseMode = "MarkdownV2"
			mm.Text = controlContrainer.GetTitle()

			bot.Send(mm)
		}

		sessionStore.Store(update.Message.Chat.ID, session)

	}
}

func sendReply(bot *tgbotapi.BotAPI, chatID int64, text string, removeKb bool) {
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = tgbotapi.ModeMarkdownV2
	if removeKb {
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	}
	_, err := bot.Send(reply)
	if err != nil {
		log.Println(err)
	}
}

type Session struct {
	BFCState   *botform.State
	BFCStarted bool
}

func loadBfc() *botform.BotFormComposer {
	return &botform.BotFormComposer{
		Elements: map[string]*botform.BotFormElementNode{
			"is_ok": &botform.BotFormElementNode{
				Renderer: botform.BotChoiceKeyboardControl{
					Type:          "BotChoiceKeyboardControl",
					Title:         "Save and exit?",
					ButtonsPerRow: 2,
				},
				DataPath: []string{"is_ok"},
				Required: false,
			},
			"responsibility": &botform.BotFormElementNode{
				Renderer: botform.BotMultiChoiceKeyboardControl{
					Type:            "BotChoiceKeyboardControl",
					Title:           "Your responsibilites?",
					ButtonsPerRow:   2,
					SelectedPrefix:  "[+] ",
					ValuesSeparator: ";",
				},
				DataPath: []string{"responsibility"},
				Required: false,
			},
			"custom_hobby": &botform.BotFormElementNode{
				Renderer: botform.BotTextKeyboardControl{
					Type:  "BotTextKeyboardControl",
					Title: "Enter your hobby",
				},
				DataPath: []string{},
				Required: false,
			},
			"hobbies": &botform.BotFormElementNode{
				Renderer: botform.BotChoiceKeyboardControl{
					Type:  "BotChoiceKeyboardControl",
					Title: "Select you hobby",
				},
				DataPath: []string{"hobbies"},
				Required: true,
			},
			"username": &botform.BotFormElementNode{
				Renderer: botform.BotTextKeyboardControl{
					Type:  "BotTextKeyboardControl",
					Title: "What is your name?",
				},
				DataPath: []string{},
				Required: true,
			},
		},
		OrderTree: &botform.BotFormDataNode{Children: map[string]*botform.BotFormDataNode{
			"*": &botform.BotFormDataNode{
				Name: "username",
				Children: map[string]*botform.BotFormDataNode{
					"*": &botform.BotFormDataNode{Name: "hobbies", Children: map[string]*botform.BotFormDataNode{
						"*": &botform.BotFormDataNode{Name: "is_ok"},
						"5": &botform.BotFormDataNode{Name: "responsibility", Children: map[string]*botform.BotFormDataNode{
							"*": &botform.BotFormDataNode{Name: "is_ok"},
						}},
						"other": &botform.BotFormDataNode{Name: "custom_hobby", Children: map[string]*botform.BotFormDataNode{"*": &botform.BotFormDataNode{Name: "is_ok"}}},
					}},
				}},
		}, Name: "root"},
		Data: &botform.BotFormDataNode{
			Name:  "root",
			Value: "",
			Children: map[string]*botform.BotFormDataNode{
				"responsibility": &botform.BotFormDataNode{
					Name: "responsibility",
					Children: map[string]*botform.BotFormDataNode{
						"a1": &botform.BotFormDataNode{Name: "a1", Value: "Clean weapons"},
						"a2": &botform.BotFormDataNode{Name: "a2", Value: "Shoot"},
						"a3": &botform.BotFormDataNode{Name: "a3", Value: "Talk to radio"},
						"a4": &botform.BotFormDataNode{Name: "a4", Value: "Not doing shit"},
						"a5": &botform.BotFormDataNode{Name: "a5", Value: "Respect Zalujniy"},
					},
				},
				"hobbies": &botform.BotFormDataNode{
					Name: "hobbies",
					Children: map[string]*botform.BotFormDataNode{
						"1":     &botform.BotFormDataNode{Name: "1", Value: "Swimming"},
						"2":     &botform.BotFormDataNode{Name: "2", Value: "Cycling"},
						"3":     &botform.BotFormDataNode{Name: "3", Value: "Running"},
						"4":     &botform.BotFormDataNode{Name: "4", Value: "Snowboard"},
						"5":     &botform.BotFormDataNode{Name: "5", Value: "Doing war"},
						"6":     &botform.BotFormDataNode{Name: "6", Value: "Videogames"},
						"7":     &botform.BotFormDataNode{Name: "7", Value: "Travelling"},
						"8":     &botform.BotFormDataNode{Name: "8", Value: "Reading"},
						"other": &botform.BotFormDataNode{Name: "other", Value: "Not in list"},
					},
				},
				"is_ok": &botform.BotFormDataNode{Name: "is_ok", Children: map[string]*botform.BotFormDataNode{
					"yes": &botform.BotFormDataNode{Name: "yes", Value: "Yes Yes Yes"},
					"no":  &botform.BotFormDataNode{Name: "no", Value: "No No No"},
				}},
			},
		},
	}
}
