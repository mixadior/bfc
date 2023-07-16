package botform

type ControlChoicesConfig struct {
	Slot1Title  string
	Slot1Action string

	Slot2Title  string
	Slot2Action string

	Slot3Title  string
	Slot3Action string
}

func (bfc *BotFormComposer) GenerateControlChoices(state *State, currentKey BfcKey) (*ControlChoicesConfig, error) {

	el, _ := bfc.GetElement(currentKey.GetLast())
	r := el.Renderer

	controlChoices := &ControlChoicesConfig{}
	isAllowedPrev, _, _ := bfc.GetPrevStepToJump(state)
	if isAllowedPrev {
		controlChoices.Slot1Title = state.PrevActionLabel
		controlChoices.Slot1Action = ControlOnInputActionPrevious
	}

	controlChoices.Slot2Title = state.CancelActionLabel
	controlChoices.Slot2Action = ControlOnInputActionCancel

	isMultiChoice := false
	switch r.(type) {
	case BotMultiChoiceKeyboardControl:
		controlChoices.Slot3Title = state.SaveAndNextActionLabel
		controlChoices.Slot3Action = ControlOnInputActionMultiSaveAndNext
		isMultiChoice = true
	default:
		controlChoices.Slot3Title = state.NextActionLabel
		controlChoices.Slot3Action = ControlOnInputActionJumpNext
	}

	isAllowedNext, _, err := bfc.GetNextToJump(state)
	if isAllowedNext {
		if len(controlChoices.Slot3Action) == 0 {
			controlChoices.Slot3Action = ControlOnInputActionJumpNext
		}
	} else {
		if err != nil {
			if err.Error() == "[next] no next element" {
				if isMultiChoice {
					controlChoices.Slot3Title = state.SaveAndNextActionLabel
				} else {
					controlChoices.Slot3Title = state.NextActionLabel
				}
				controlChoices.Slot3Action = ControlOnInputActionFinish
			} else {
				return controlChoices, err
			}
		} else {
			if isMultiChoice {
				multiChoiceValues, ok := state.Values[currentKey.GetLast()]
				if !ok || len(multiChoiceValues) == 0 {
					controlChoices.Slot3Title = ""
				}
			} else {
				controlChoices.Slot3Title = ""
			}
		}
	}

	return controlChoices, nil
}

func (bfc *BotFormComposer) CheckControlChoice(state *State, currentKey BfcKey, value string) (action string) {
	choicesConfig, err := bfc.GenerateControlChoices(state, currentKey)
	if err != nil {
		return ControlOnInputActionError
	}
	switch true {
	case choicesConfig.Slot1Title == value:
		return choicesConfig.Slot1Action
	case choicesConfig.Slot2Title == value:
		return choicesConfig.Slot2Action
	case choicesConfig.Slot3Title == value:
		return choicesConfig.Slot3Action
	}

	return ""
}
