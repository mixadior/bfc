package main

import (
	"context"
	"fmt"
	botform "github.com/mixadior/bfc"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"github.com/mr-linch/go-tg/tgb/session"
	"os"
	"os/signal"
	"syscall"
)

type Session struct {
	BFCState     *botform.State
	BfcIsStarted bool
}

func main() {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		defer os.Exit(1)
	}
}

func run(ctx context.Context) error {
	client := tg.New(os.Getenv("BOT_TOKEN"))
	sm := session.NewManager(&Session{
		BFCState:     botform.NewState(map[string]string{}, botform.BfcKey{"root"}),
		BfcIsStarted: false,
	}, session.WithStore(session.NewStoreMemory()))

	bfc := loadBfc()
	router := tgb.NewRouter().Use(sm). // handles /start and /help
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer(
				tg.HTML.Text(
					tg.HTML.Bold("👋 Hi, I'm echo bot!"),
					"",
					tg.HTML.Italic("🚀 Powered by", tg.HTML.Spoiler("mr-linch/go-tg")),
					tg.HTML.Line(tg.HTML.Text("Type /wizard to check BFC.")),
				),
			).ParseMode(tg.HTML).DoVoid(ctx)
		}, tgb.Command("start", tgb.WithCommandAlias("help"))).
		// handle other messages
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			messageText := msg.Text

			state := (*sm.Get(ctx)).BFCState
			if messageText == "/wizard" {
				state = botform.NewState(map[string]string{}, botform.BfcKey{"root"})
				(*sm.Get(ctx)).BFCState = state
				(*sm.Get(ctx)).BfcIsStarted = true
			}

			if !(*sm.Get(ctx)).BfcIsStarted {
				return msg.Answer("I donot know. Try /wizard to test bfc.").DoVoid(ctx)
			}

			validationMessage := ""
			v, e, nextAction := bfc.ValidateAndTransform(state, messageText)
			if e != nil {
				validationMessage = e.Error()
				msg.Answer(validationMessage).DoVoid(ctx)
			} else {
				bfc.CommitToState(v, nextAction, state)
			}

			if state.IsFinished {
				isOk := state.Values["is_ok"]
				if isOk == "yes" {
					prettyValues := bfc.GetPrettyValues(state)
					texts := []string{}
					for keyTitle, keyValue := range prettyValues {
						texts = append(texts, tg.MD2.Line(tg.MD2.Bold(tg.MD2.Escape(keyTitle)), tg.MD2.Text(tg.MD2.Escape(keyValue))))
					}
					msg.Answer("Answers saved...").DoVoid(ctx)
					e := msg.Answer(tg.MD2.Text(texts...)).ParseMode(tg.MD2).ReplyMarkup(tg.NewReplyKeyboardRemove()).DoVoid(ctx)
					fmt.Println(e)
					(*sm.Get(ctx)).BfcIsStarted = false

				} else {
					msg.Answer("Answers not saved. Thanks.").ReplyMarkup(tg.NewReplyKeyboardRemove()).DoVoid(ctx)
				}
			} else if state.IsCanceled {
				(*sm.Get(ctx)).BfcIsStarted = false
				msg.Answer("Canceled.").ReplyMarkup(tg.NewReplyKeyboardRemove()).DoVoid(ctx)
			} else {
				controlContrainer, _ := bfc.Build(state)

				switch controlContrainer.(type) {
				case botform.BotTextKeyboardControl:
					bckc := controlContrainer.(botform.BotTextKeyboardControl)
					rows := [][]tg.KeyboardButton{}

					controlRow := []tg.KeyboardButton{}
					ssmChoices := botform.SortedStringMap(bckc.ControlChoices)
					ssmChoices.Range(func(k, choice string) bool {
						controlRow = append(controlRow, tg.NewKeyboardButton(choice))
						return true
					}, false)

					if len(controlRow) > 0 {
						rows = append(rows, controlRow)
					}
					msg.Answer(controlContrainer.GetTitle()).ReplyMarkup(tg.NewReplyKeyboardMarkup(rows...)).DoVoid(ctx)

				case botform.BotChoiceKeyboardControl:
					bckc := controlContrainer.(botform.BotChoiceKeyboardControl)
					rows := [][]tg.KeyboardButton{}

					controlRow := []tg.KeyboardButton{}
					ssmChoices := botform.SortedStringMap(bckc.ControlChoices)
					ssmChoices.Range(func(k, choice string) bool {
						controlRow = append(controlRow, tg.NewKeyboardButton(choice))
						return true
					}, false)

					if len(controlRow) > 0 {
						rows = append(rows, controlRow)
					}

					row := []tg.KeyboardButton{}
					ssmMainChoices := botform.SortedStringMap(bckc.Choices)
					ssmMainChoices.Range(func(k, choice string) bool {
						row = append(row, tg.NewKeyboardButton(choice))
						if len(row) >= bckc.ButtonsPerRow {
							rows = append(rows, row)
							row = []tg.KeyboardButton{}
						}
						return true
					}, true)

					if len(row) > 0 {
						rows = append(rows, row)
					}
					msg.Answer(controlContrainer.GetTitle()).ReplyMarkup(tg.NewReplyKeyboardMarkup(rows...)).DoVoid(ctx)
				default:
					msg.Answer(controlContrainer.GetTitle()).DoVoid(ctx)
				case botform.BotMultiChoiceKeyboardControl:
					bckc := controlContrainer.(botform.BotMultiChoiceKeyboardControl)
					rows := [][]tg.KeyboardButton{}

					controlRow := []tg.KeyboardButton{}
					ssmChoices := botform.SortedStringMap(bckc.ControlChoices)
					ssmChoices.Range(func(k, choice string) bool {
						controlRow = append(controlRow, tg.NewKeyboardButton(choice))
						return true
					}, false)

					if len(controlRow) > 0 {
						rows = append(rows, controlRow)
					}

					row := []tg.KeyboardButton{}
					ssmMainChoices := botform.SortedStringMap(bckc.Choices)
					ssmMainChoices.Range(func(k, choice string) bool {
						_, existsInSelected := bckc.SelectedValues[choice]
						if existsInSelected {
							choice = bckc.SelectedPrefix + choice
						}
						row = append(row, tg.NewKeyboardButton(choice))
						if len(row) >= bckc.ButtonsPerRow {
							rows = append(rows, row)
							row = []tg.KeyboardButton{}
						}
						return true
					}, true)

					if len(row) > 0 {
						rows = append(rows, row)
					}
					msg.Answer(controlContrainer.GetTitle()).ReplyMarkup(tg.NewReplyKeyboardMarkup(rows...)).DoVoid(ctx)
				}
				//fmt.Println(controlContrainer.GetTitle(), validationMessage)

			}

			return e
		})

	return tgb.NewPoller(
		router,
		client,
	).Run(ctx)
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
			"country": &botform.BotFormElementNode{
				Renderer: botform.BotChoiceKeyboardControl{
					Type:          "BotChoiceKeyboardControl",
					Title:         "Country",
					ButtonsPerRow: 2,
				},
				DataPath: []string{"country"},
				Required: false,
			},
			"state": &botform.BotFormElementNode{
				Renderer: botform.BotChoiceKeyboardControl{
					Type:          "BotChoiceKeyboardControl",
					Title:         "State",
					ButtonsPerRow: 2,
				},
				DataPath: []string{"country", "state"},
				Required: false,
			},
			"state_city": &botform.BotFormElementNode{
				Renderer: botform.BotChoiceKeyboardControl{
					Type:          "BotChoiceKeyboardControl",
					Title:         "State city",
					ButtonsPerRow: 2,
				},
				DataPath: []string{"country", "state", "state_city"},
				Required: false,
			},
			"custom_city": &botform.BotFormElementNode{
				Renderer: botform.BotTextKeyboardControl{
					Type:  "BotTextKeyboardControl",
					Title: "City",
				},
				DataPath: []string{},
				Required: true,
			},
			"countries_to_add": &botform.BotFormElementNode{
				Renderer: botform.BotMultiChoiceKeyboardControl{
					Type:            "BotMultiChoiceKeyboardControl",
					Title:           "Choose countries, what you want we should add.",
					SelectedPrefix:  "[+] ",
					ValuesSeparator: ";",
				},
				DataPath: []string{"countries_to_add"},
				Required: true,
			},
		},
		OrderTree: &botform.BotFormDataNode{Children: map[string]*botform.BotFormDataNode{
			"*": &botform.BotFormDataNode{
				Name: "country",
				Children: map[string]*botform.BotFormDataNode{
					"not_in_list": &botform.BotFormDataNode{Name: "countries_to_add", Children: map[string]*botform.BotFormDataNode{"*": &botform.BotFormDataNode{Name: "is_ok"}}},
					"usa": &botform.BotFormDataNode{
						Name: "state",
						Children: map[string]*botform.BotFormDataNode{
							"*": &botform.BotFormDataNode{Name: "is_ok"},
							"texas": &botform.BotFormDataNode{
								Name: "state_city",
								Children: map[string]*botform.BotFormDataNode{
									"whole_state": &botform.BotFormDataNode{Name: "is_ok"},
									"custom_city": &botform.BotFormDataNode{Name: "custom_city", Children: map[string]*botform.BotFormDataNode{"*": &botform.BotFormDataNode{Name: "is_ok"}}},
									"*":           &botform.BotFormDataNode{Name: "is_ok"},
								},
							},
						},
					},
				}},
		}, Name: "root"},
		Data: &botform.BotFormDataNode{
			Name:  "root",
			Value: "",
			Children: map[string]*botform.BotFormDataNode{
				"countries_to_add": &botform.BotFormDataNode{
					Name: "countries_to_add",
					Children: map[string]*botform.BotFormDataNode{
						"ua": &botform.BotFormDataNode{Name: "ua", Value: "Ukraine"},
						"pl": &botform.BotFormDataNode{Name: "pl", Value: "Poland"},
						"it": &botform.BotFormDataNode{Name: "it", Value: "Italy"},
						"es": &botform.BotFormDataNode{Name: "es", Value: "Spain"},
						"pt": &botform.BotFormDataNode{Name: "pt", Value: "Portugal"},
						"gr": &botform.BotFormDataNode{Name: "gr", Value: "Greece"},
						"ee": &botform.BotFormDataNode{Name: "ee", Value: "Estonia"},
						"de": &botform.BotFormDataNode{Name: "de", Value: "Germany"},
					},
				},
				"is_ok": &botform.BotFormDataNode{Name: "is_ok", Children: map[string]*botform.BotFormDataNode{
					"yes": &botform.BotFormDataNode{Name: "yes", Value: "Yes Yes Yes"},
					"no":  &botform.BotFormDataNode{Name: "no", Value: "No No No"},
				}},
				"country": &botform.BotFormDataNode{
					Name:  "country",
					Value: "Country",
					Children: map[string]*botform.BotFormDataNode{
						"not_in_list": &botform.BotFormDataNode{Value: "Not in list...", Name: "not_in_list"},
						"usa": &botform.BotFormDataNode{
							Name:  "usa",
							Value: "United states of America",
							Children: map[string]*botform.BotFormDataNode{
								"alabama": &botform.BotFormDataNode{Name: "alabama", Value: "Alabama"},
								"texas": &botform.BotFormDataNode{
									Name:  "texas",
									Value: "Texas",
									Children: map[string]*botform.BotFormDataNode{
										"whole_state": &botform.BotFormDataNode{
											Name:  "whole_state",
											Value: "Whole state",
										},
										"custom_city": &botform.BotFormDataNode{
											Name:  "custom_city",
											Value: "Custom city",
										},
										"753": &botform.BotFormDataNode{Name: "753", Value: "Dallas"},
										"752": &botform.BotFormDataNode{Name: "752", Value: "Austin"},
										"751": &botform.BotFormDataNode{Name: "751", Value: "Houston"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
