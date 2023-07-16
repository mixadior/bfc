package botform

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ControlOnInputActionRerender         string = "current"
	ControlOnInputActionJumpNext                = "jumpnext"
	ControlOnInputActionNext                    = "next"
	ControlOnInputActionMultiSaveAndNext        = "multisavenext"
	ControlOnInputActionError                   = "error"
	ControlOnInputActionPrevious                = "prev"
	ControlOnInputActionCancel                  = "cancel"
	ControlOnInputActionFinish                  = "finish"
)

type BotFormComposer struct {
	Data      *BotFormDataNode
	OrderTree *BotFormDataNode
	Elements  map[string]*BotFormElementNode
}

func (bfc *BotFormComposer) GetNext(state *State, forKey BfcKey) (string, bool) {
	node := bfc.OrderTree
	for _, kk := range forKey {
		fieldValue, okValue := state.Values[kk]
		nodeByValue, okByValue := node.Children[fieldValue]
		if okByValue {
			node = nodeByValue
			continue
		}

		_, okAny := node.Children["*"]
		if okAny {
			node = node.Children["*"]
		} else {
			if len(node.Children) > 0 {
				return "", true
			} else if !okValue {
				return fieldValue, okValue
			}

			if node.Children == nil || len(node.Children) == 0 {
				return "", false
			}
			//node = node.Children[fieldValue]
		}
	}
	if node == nil {
		return "", false
	}
	return node.Name, true
}

func (bfc *BotFormComposer) GetElement(forKey string) (*BotFormElementNode, bool) {
	v, ok := bfc.Elements[forKey]
	return v, ok
}

func (bfc *BotFormComposer) GetNextToJump(state *State) (isAllowed bool, nextKey string, err error) {
	isAllowed = true
	elementKey := state.Key.GetLast()
	forKey := state.Key

	if elementKey == "root" {
		nextKey, _ = bfc.GetNext(state, state.Key)
		return true, nextKey, nil
	}

	currentElement, ok := bfc.GetElement(elementKey)
	if !ok {
		err = errors.New("[next] cannot find current element")
	} else {
		if currentElement.Required {
			isAllowed = false
		} else {
		nextKeySearch:
			for {
				nextKey, ok = bfc.GetNext(state, forKey)
				if !ok && len(nextKey) == 0 {
					isAllowed = false
					err = errors.New("[next] no next element")
					break
				} else if len(nextKey) == 0 && ok {
					isAllowed = false
					break
				}
				nextElement, ok := bfc.GetElement(nextKey)
				if ok {
					// we check if forKey is required for next element, checking it in datapath
					for _, v := range nextElement.DataPath {
						if v == forKey.GetLast() {
							forKey = append(forKey, nextKey)
							continue nextKeySearch
						}
					}
					break
				} else {
					err = errors.New("[next] no next element")
					break
				}
			}
		}
	}
	return isAllowed, nextKey, err
}

func (bfc *BotFormComposer) GetPrev(forKey BfcKey) (string, bool) {
	prevKey := len(forKey) - 2
	if prevKey >= 0 {
		return forKey[prevKey], true
	} else {
		return "", false
	}
}

func (bfc *BotFormComposer) GetPrevStepToJump(state *State) (isAllowed bool, prevKey string, err error) {
	isAllowed = true
	forKey := state.Key
	if forKey.GetLast() == "root" {
		isAllowed = false
		prevKey = ""
	} else {
		// previous key with value or root
		for {
			var ok bool
			prevKey, ok = bfc.GetPrev(forKey)
			if prevKey == "root" {
				isAllowed = false
				break
			}
			if !ok {
				err = errors.New("[prev] cannot find node")
				isAllowed = false
				break
			} else {
				_, ok := state.Values[prevKey]
				if ok {
					break
				} else {
					// we allow if this element is root of some branch
					el, ok := bfc.GetElement(prevKey)
					if !ok {
						err = errors.New("[prev] no element")
						break
					}
					if len(el.DataPath) <= 1 {
						break
					}
					forKey = forKey.Shift()
				}
			}
		}
	}
	return isAllowed, prevKey, err
}

func (bfc *BotFormComposer) ValidateAndTransform(state *State, rawValue string) (newValue string, e error, nextAction string) {
	currentKey := state.Key
	element := currentKey.GetLast()
	if element == "root" {
		// if not control and is root - skip
		return rawValue, nil, ControlOnInputActionNext
	}

	// check control messages
	controlAction := bfc.CheckControlChoice(state, currentKey, rawValue)
	if controlAction != "" {
		return "is_control_choice", nil, controlAction
	}

	elementO, ok := bfc.GetElement(element)
	if !ok {
		return "", errors.New(fmt.Sprintf("No such element: %s", element)), ControlOnInputActionError
	}
	switch elementO.Renderer.(type) {
	case BotMultiChoiceKeyboardControl:
		realValues := []string{}

		renderer := elementO.Renderer.(BotMultiChoiceKeyboardControl)
		renderer.SelectedValues = map[string]string{}

		elV, ok := state.Values[element]
		parsedValues := []string{}
		if ok {

			realValues = strings.Split(elV, renderer.ValuesSeparator)
			parsedValues = make([]string, len(realValues))
			choices, e := bfc.GetChoicesByPath(state, elementO.DataPath, false)
			if e != nil {
				return "", e, ControlOnInputActionError
			}

			for k, v := range realValues {
				v, ok := choices[v]
				if ok {
					parsedValues[k] = v
				}
			}
		}

		if strings.HasPrefix(rawValue, renderer.SelectedPrefix) {
			unprefixedValue := strings.TrimPrefix(rawValue, renderer.SelectedPrefix)
			removeKey := -1
			for k, v := range parsedValues {
				if unprefixedValue == v {
					removeKey = k
					break
				}
			}
			if removeKey >= 0 {
				realValues = RemoveIndex(realValues, removeKey)
			}

		} else {
			rv, e := bfc.GetElementValueByText(state, element, rawValue)
			if e != nil {
				return "", e, ControlOnInputActionError
			}
			realValues = append(realValues, rv)
		}

		emptyIndexes := []int{}
		for k, v := range realValues {
			if len(strings.TrimSpace(v)) == 0 {
				emptyIndexes = append(emptyIndexes, k)
			}
		}

		for _, v := range emptyIndexes {
			realValues = RemoveIndex(realValues, v)
		}
		return strings.Join(realValues, renderer.ValuesSeparator), nil, ControlOnInputActionRerender

	case BotTextKeyboardControl:
		return rawValue, nil, ControlOnInputActionNext
	case BotChoiceKeyboardControl:
		v, e := bfc.GetElementValueByText(state, element, rawValue)
		if e != nil {
			return "", e, ControlOnInputActionError
		}

		return v, nil, ControlOnInputActionNext
	}
	return "", errors.New("cannot validate data"), ControlOnInputActionNext
}

func (bfc *BotFormComposer) GetElementValueByText(state *State, element string, txt string) (string, error) {
	el, ok := bfc.GetElement(element)
	if !ok {
		return "", errors.New("")
	}
	currentChoices, e := bfc.GetChoicesByPath(state, el.DataPath, false)
	if e != nil {
		return "", e
	}
	txt = strings.TrimSpace(txt)
	for v, curChoice := range currentChoices {
		if txt == strings.TrimSpace(curChoice) {
			return v, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Didnot find key %s for element %s", txt, element))
}

func (bfc *BotFormComposer) GetChoicesByPath(state *State, path []string, skipHidden bool) (map[string]string, error) {
	node := bfc.Data.Children[path[0]]

	if len(path) > 1 {
		for kk, pv := range path {

			if kk == (len(path) - 1) {
				break
			}

			vv, ok := state.Values[pv]
			if !ok {
				return nil, errors.New(fmt.Sprintf("Value doesnot exists: %s", strings.Join(path, "->")))
			}

			node, ok = node.Children[vv]
			if !ok {
				return nil, errors.New(fmt.Sprintf("Path doesnot exists: %s", strings.Join(path, "->")))
			}
		}
	}

	choices := map[string]string{}
	for k, v := range node.Children {
		if skipHidden && v.Hidden {
			continue
		}
		choices[k] = v.Value
	}
	return choices, nil
}

func (bfc *BotFormComposer) GetPrettyValues(state *State) map[string]string {

	prettyValues := map[string]string{}
	currentValues := state.GetValues()
	for elementName, dataValue := range currentValues {
		el, okEl := bfc.GetElement(elementName)
		if !okEl {
			continue
		}
		var prettyValue string
		switch el.Renderer.(type) {
		default:
			prettyValue = dataValue
		case BotMultiChoiceKeyboardControl:
			ell := el.Renderer.(BotMultiChoiceKeyboardControl)
			choosenValues := strings.Split(dataValue, ell.ValuesSeparator)
			choices, e := bfc.GetChoicesByPath(state, el.DataPath, false)
			if e != nil {
				prettyValue = dataValue
				continue
			}
			prettyValues := []string{}
			for _, v := range choosenValues {
				prettyValueChoice, ok := choices[v]
				if !ok {
					prettyValues = append(prettyValues, v)
				} else {
					prettyValues = append(prettyValues, prettyValueChoice)
				}
			}
			prettyValue = strings.Join(prettyValues, ", ")
		case BotTextKeyboardControl:
			prettyValue = dataValue
		case BotChoiceKeyboardControl:
			choices, e := bfc.GetChoicesByPath(state, el.DataPath, false)
			if e != nil {
				continue
			}
			prettyValue = choices[dataValue]
		}

		prettyKey := el.Renderer.GetTitle()
		prettyValues[prettyKey] = prettyValue
	}
	return prettyValues
}

func (bfc *BotFormComposer) CommitToState(validatedValue string, action string, state *State) (e error) {

	keyToUpdate := state.Key.GetLast()
	if validatedValue != "is_control_choice" {
		state.Store(keyToUpdate, validatedValue)
	}

	switch action {
	case ControlOnInputActionCancel:
		state.IsCanceled = true
	case ControlOnInputActionRerender:

	case ControlOnInputActionPrevious:
		_, prevStepString, _ := bfc.GetPrevStepToJump(state)
		// delete all values after step that selected to jump
		prevStepKey := 0

		deleteKeys := []string{}
		deleteKeysDo := false
		for k, currentStepStep := range state.Key {
			if currentStepStep == prevStepString && !deleteKeysDo {
				prevStepKey = k
				deleteKeysDo = true
			}

			if deleteKeysDo {
				deleteKeys = append(deleteKeys, currentStepStep)
			}
		}

		for _, deleteKey := range deleteKeys {
			state.Delete(deleteKey)
		}

		if prevStepKey > 0 {
			state.Key = state.Key[:prevStepKey+1]
		}

	case ControlOnInputActionNext, ControlOnInputActionMultiSaveAndNext:
		currentElementKeyString, hasNext := bfc.GetNext(state, state.Key)
		if hasNext {
			state.Key = append(state.Key, currentElementKeyString)
		} else {
			state.IsFinished = true
		}
	case ControlOnInputActionJumpNext:
		hasNext, currentElementKeyString, err := bfc.GetNextToJump(state)
		if hasNext {
			state.Key = append(state.Key, currentElementKeyString)
		} else {
			state.IsFinished = true
		}
		e = err
	case ControlOnInputActionFinish:
		state.IsFinished = true
	}

	return e
}

func (bfc *BotFormComposer) Build(state *State) (BotControl, error) {

	forKey := state.Key
	el, _ := bfc.GetElement(forKey.GetLast())
	r := el.Renderer

	controlChoices := map[string]string{}
	controlChoicesConfig, _ := bfc.GenerateControlChoices(state, forKey)

	if controlChoicesConfig.Slot1Action != "" {
		controlChoices["a"] = controlChoicesConfig.Slot1Title
	}
	if controlChoicesConfig.Slot2Action != "" {
		controlChoices["b"] = controlChoicesConfig.Slot2Title
	}
	if controlChoicesConfig.Slot3Action != "" {
		controlChoices["c"] = controlChoicesConfig.Slot3Title
	}

	switch r.(type) {
	case BotTextKeyboardControl:
		rr := r.(BotTextKeyboardControl)
		rr.SetControlChoices(controlChoices)
		return rr, nil
	case BotMultiChoiceKeyboardControl:
		rr := r.(BotMultiChoiceKeyboardControl)
		choices, err := bfc.GetChoicesByPath(state, el.DataPath, true)
		if err != nil {
			return nil, err
		}
		rr.SetChoices(choices)
		rr.SetControlChoices(controlChoices)
		v, ok := state.Values[forKey.GetLast()]
		if ok {
			parsedV := strings.Split(v, rr.ValuesSeparator)
			selectedValues := map[string]string{}
			choices, err := bfc.GetChoicesByPath(state, el.DataPath, false)
			if err == nil {
				for _, pv := range parsedV {
					choiceText, ok := choices[pv]
					if ok {
						selectedValues[choiceText] = choiceText
					}
				}
			}
			rr.SelectedValues = selectedValues
		}
		return rr, nil
	case BotChoiceKeyboardControl:
		rr := r.(BotChoiceKeyboardControl)
		choices, err := bfc.GetChoicesByPath(state, el.DataPath, true)
		if err != nil {
			return nil, err
		}
		rr.SetChoices(choices)
		rr.SetControlChoices(controlChoices)
		return rr, nil
	}

	return r, nil
}
