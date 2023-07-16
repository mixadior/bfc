package main

import (
	"context"
	_ "embed"
	"fmt"
	botform "github.com/mixadior/bfc"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"github.com/mr-linch/go-tg/tgb/session"
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"syscall"
)

type Session struct {
	BFCState     *botform.State
	BfcIsStarted bool
}

//go:embed olxdata/out_data.yml
var olxData []byte

//go:embed olxdata/out_elements_filters.yml
var elementsDataData []byte

//go:embed olxdata/out_data_order_tree.yml
var cityDataNodeOrderData []byte

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

	cityDataNode := &botform.BotFormDataNode{}
	yaml.Unmarshal(cityDataNodeOrderData, cityDataNode)

	olxDataNode := &botform.BotFormDataNode{}
	yaml.Unmarshal(olxData, olxDataNode)

	olxElementsNode := map[string]*botform.BotFormElementNode{}
	yaml.Unmarshal(elementsDataData, olxElementsNode)

	elements := olxElementsNode

	elements["region"] = &botform.BotFormElementNode{
		Renderer: botform.BotChoiceKeyboardControl{
			Title:         "Область",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region"},
	}
	elements["city"] = &botform.BotFormElementNode{
		Renderer: botform.BotChoiceKeyboardControl{
			Type:          "BotMultiChoiceKeyboardControl",
			Title:         "Місто",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region", "city"},
	}

	elements["district"] = &botform.BotFormElementNode{
		Renderer: botform.BotChoiceKeyboardControl{
			Title:         "Район міста",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region", "city", "district"},
	}

	elements["currency"].Required = true

	bfc := &botform.BotFormComposer{
		OrderTree: &botform.BotFormDataNode{Children: map[string]*botform.BotFormDataNode{
			"*": &botform.BotFormDataNode{
				Name: "region",
				Children: map[string]*botform.BotFormDataNode{
					"*": cityDataNode,
				}},
		}, Name: "root"},
		Data:     olxDataNode,
		Elements: elements,
	}

	return bfc
}
