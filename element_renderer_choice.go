package botform

type BotChoiceKeyboardControl struct {
	Type           string
	Title          string
	Choices        map[string]string
	ControlChoices map[string]string
	ButtonsPerRow  int
	Selected       string
}

func (bckc BotChoiceKeyboardControl) GetTitle() string {
	return bckc.Title
}

func (bckc *BotChoiceKeyboardControl) SetSelectedValue(selected string) {
	bckc.Selected = selected
}

func (bckc *BotChoiceKeyboardControl) SetChoices(c map[string]string) {
	bckc.Choices = c
}

func (bckc *BotChoiceKeyboardControl) SetControlChoices(c map[string]string) {
	bckc.ControlChoices = c
}
