package botform

type BotMultiChoiceKeyboardControl struct {
	Type            string
	Title           string
	Choices         map[string]string
	ControlChoices  map[string]string
	ButtonsPerRow   int
	ValuesSeparator string
	SelectedPrefix  string
	SelectedValues  map[string]string
}

func (bckc BotMultiChoiceKeyboardControl) GetTitle() string {
	return bckc.Title
}

func (bckc *BotMultiChoiceKeyboardControl) SetChoices(c map[string]string) {
	bckc.Choices = c
}

func (bckc *BotMultiChoiceKeyboardControl) SetControlChoices(c map[string]string) {
	bckc.ControlChoices = c
}
