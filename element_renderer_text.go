package botform

type BotTextKeyboardControl struct {
	Type           string
	Title          string
	ControlChoices map[string]string
	Selected       string
}

func (bckc BotTextKeyboardControl) GetTitle() string {
	return bckc.Title
}

func (bckc *BotTextKeyboardControl) SetControlChoices(c map[string]string) {
	bckc.ControlChoices = c
}
