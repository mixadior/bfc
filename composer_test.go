package botform

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func loadBfc(subdir string) *BotFormComposer {

	outData, _ := os.ReadFile(filepath.Join("testdata", subdir, "out_data.yml"))
	outOrders, _ := os.ReadFile(filepath.Join("testdata", subdir, "out_data_order_tree.yml"))
	outFilters, _ := os.ReadFile(filepath.Join("testdata", subdir, "out_elements_filters.yml"))

	olxDataNode := &BotFormDataNode{}
	yErr := yaml.Unmarshal(outData, olxDataNode)
	if yErr != nil {

	}

	olxElementsNode := map[string]*BotFormElementNode{}
	yErr = yaml.Unmarshal(outFilters, olxElementsNode)
	if yErr != nil {
		fmt.Println(yErr)
		os.Exit(0)
	}

	cityDataNode := &BotFormDataNode{}
	yaml.Unmarshal(outOrders, cityDataNode)

	olxElementsNode["region"] = &BotFormElementNode{
		Renderer: BotChoiceKeyboardControl{
			Title:         "Область",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region"},
	}
	olxElementsNode["city"] = &BotFormElementNode{
		Renderer: BotChoiceKeyboardControl{
			Title:         "Місто",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region", "city"},
	}

	olxElementsNode["district"] = &BotFormElementNode{
		Renderer: BotChoiceKeyboardControl{
			Title:         "Район міста",
			ButtonsPerRow: 2,
		},
		DataPath: []string{"region", "city", "district"},
	}

	olxElementsNode["currency"].Required = true

	bfc := &BotFormComposer{
		OrderTree: &BotFormDataNode{Children: map[string]*BotFormDataNode{
			"*": &BotFormDataNode{
				Name: "region",
				Children: map[string]*BotFormDataNode{
					"*": cityDataNode,
				}},
		}, Name: "root"},
		Data:     olxDataNode,
		Elements: olxElementsNode,
	}
	return bfc
}

type StepTestData struct {
	Value          string
	ActionShouldBe string
	StepAfter      string
}

func NewStepTestData(value, action, stepafter string) StepTestData {
	return StepTestData{value, action, stepafter}
}

func TestBfcSteps(t *testing.T) {
	bfc := loadBfc("pointer_remove")
	steps := []StepTestData{
		NewStepTestData("/start", ControlOnInputActionNext, "region"),
		NewStepTestData("Вінницька область", ControlOnInputActionNext, "city"),
		NewStepTestData("=>", ControlOnInputActionJumpNext, "currency"),
		NewStepTestData("=>", ControlOnInputActionError, "currency"),
		NewStepTestData("<=", ControlOnInputActionPrevious, "region"),
		NewStepTestData("Вінницька область", ControlOnInputActionNext, "city"),
		NewStepTestData("Вінниця", ControlOnInputActionNext, "district"),
		NewStepTestData("=>", ControlOnInputActionJumpNext, "currency"),
		NewStepTestData("грн.", ControlOnInputActionNext, "category_1"),
		NewStepTestData("Будинки", ControlOnInputActionNext, "category_206"),
		NewStepTestData("=>", ControlOnInputActionError, "category_206"),
		NewStepTestData("Продаж будинків", ControlOnInputActionNext, "filter_float_price:from"),
	}

	state := NewState(map[string]string{}, BfcKey{"root"})
	state.NextActionLabel = "=>"
	state.SaveAndNextActionLabel = "=>"

	for stepKey, step := range steps {
		newValue, e, nextAction := bfc.ValidateAndTransform(state, step.Value)
		if e != nil {
			if step.ActionShouldBe != ControlOnInputActionError {
				t.Errorf("validate and transform shouldn't error: %v", e)
				break
			}
		}

		stepActionShouldBe := step.ActionShouldBe
		if stepActionShouldBe != nextAction {
			t.Errorf("Step %d with value \"%s\" should have action \"%s\", but have \"%s\"", stepKey, step.Value, stepActionShouldBe, nextAction)
		}
		bfc.CommitToState(newValue, nextAction, state)
		stepAfter := state.Key.GetLast()
		if stepAfter != step.StepAfter {
			t.Errorf("Step after step %d with value \"%s\" should be \"%s\", but \"%s\"", stepKey, step.Value, step.StepAfter, stepAfter)
		}

		if state.IsFinished {

		}
	}
}

func TestBfcSharedMT(t *testing.T) {
	bfc := loadBfc("pointer_remove")

	state1 := NewState(map[string]string{}, BfcKey{"root"})
	state2 := NewState(map[string]string{}, BfcKey{"root"})

	wg := sync.WaitGroup{}
	results := make(chan BotControl, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		steps1 := []StepTestData{
			NewStepTestData("/start", ControlOnInputActionNext, "region"),
			NewStepTestData("Вінницька область", ControlOnInputActionNext, "city"),
		}

		for _, step := range steps1 {
			newValue, _, nextAction := bfc.ValidateAndTransform(state1, step.Value)
			bfc.CommitToState(newValue, nextAction, state1)
		}
		fe, _ := bfc.Build(state1)
		results <- fe
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		steps1 := []StepTestData{
			NewStepTestData("/start", ControlOnInputActionNext, "region"),
			NewStepTestData("Київська область", ControlOnInputActionNext, "city"),
		}

		for _, step := range steps1 {
			newValue, _, nextAction := bfc.ValidateAndTransform(state2, step.Value)
			bfc.CommitToState(newValue, nextAction, state2)
		}
		fe, _ := bfc.Build(state2)
		results <- fe
	}()

	wg.Wait()
	r1 := <-results
	r2 := <-results

	rr1 := r1.(BotChoiceKeyboardControl)
	rr2 := r2.(BotChoiceKeyboardControl)

	ssm1 := SortedStringMap(rr1.Choices)
	ssm2 := SortedStringMap(rr2.Choices)

	if len(ssm1) == 0 {
		t.Errorf("Error")
	}

	if IsEqualSST(ssm1, ssm2) {
		t.Errorf("Error")
	}

	if rr1.GetTitle() != "Місто" {
		t.Errorf("Error")
	}

}

/**
The idea of this test is to check if every session correcly copies data, we change same control values in separate
goroutines and check state's data to equality, if data is equal - means that somewhere we
*/
func TestBfcSharedMultiChoice(t *testing.T) {
	bfc := loadBfc("pointer_remove")
	steps := []StepTestData{
		NewStepTestData("/start", ControlOnInputActionNext, "region"),
		NewStepTestData("Вінницька область", ControlOnInputActionNext, "city"),
		NewStepTestData("Бар", ControlOnInputActionNext, ""),
		NewStepTestData("грн.", ControlOnInputActionNext, ""),
		NewStepTestData("Будинки", ControlOnInputActionNext, ""),
		NewStepTestData("Продаж будинків", ControlOnInputActionNext, ""),
		NewStepTestData("Skip step", ControlOnInputActionJumpNext, ""),
		NewStepTestData("Skip step", ControlOnInputActionJumpNext, ""),
		NewStepTestData("Skip step", ControlOnInputActionJumpNext, ""),
		NewStepTestData("Skip step", ControlOnInputActionJumpNext, ""),
		NewStepTestData("Skip step", ControlOnInputActionJumpNext, ""),
	}

	state1 := NewState(map[string]string{}, BfcKey{"root"})
	state2 := NewState(map[string]string{}, BfcKey{"root"})

	wg := sync.WaitGroup{}

	stepControlChan1 := make(chan BotControl)
	stepControlChan2 := make(chan BotControl)

	syncChan := make(chan bool)

	for _, step := range steps {
		newValue, _, nextAction := bfc.ValidateAndTransform(state1, step.Value)
		bfc.CommitToState(newValue, nextAction, state1)
	}

	for _, step := range steps {
		newValue, _, nextAction := bfc.ValidateAndTransform(state2, step.Value)
		bfc.CommitToState(newValue, nextAction, state2)
	}

	wg.Add(1)
	go func(s *State) {
		defer wg.Done()

		steps1 := []StepTestData{
			NewStepTestData("Індивідуальне газове", ControlOnInputActionNext, ""),
			NewStepTestData("Централізоване", ControlOnInputActionNext, ""),
		}

		for k, step := range steps1 {
			fmt.Println("cycle1", k)
			newValue, e, nextAction := bfc.ValidateAndTransform(state1, step.Value)
			if e == nil {
				bfc.CommitToState(newValue, nextAction, state1)
			}
			r, _ := bfc.Build(state1)
			stepControlChan1 <- r
			<-syncChan
		}
		close(stepControlChan1)
	}(state1)

	selectedPrefix := bfc.Elements["filter_enum_heating"].Renderer.(BotMultiChoiceKeyboardControl).SelectedPrefix
	wg.Add(1)
	go func(s *State) {
		defer wg.Done()

		steps1 := []StepTestData{
			NewStepTestData("Власна котельня", ControlOnInputActionNext, ""),
			NewStepTestData("Централізоване", ControlOnInputActionNext, ""),
			NewStepTestData(selectedPrefix+"Централізоване", ControlOnInputActionNext, ""),
		}

		for k, step := range steps1 {
			fmt.Println("cycle2", k)
			newValue, e, nextAction := bfc.ValidateAndTransform(state2, step.Value)
			if e == nil {
				bfc.CommitToState(newValue, nextAction, state2)
			}
			r, _ := bfc.Build(state2)
			stepControlChan2 <- r
			<-syncChan
		}
		close(stepControlChan2)
	}(state2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {

			vv1, v1 := <-stepControlChan1
			vv2, v2 := <-stepControlChan2

			sst1 := SortedStringMap(state1.Values)
			sst2 := SortedStringMap(state2.Values)

			if IsEqualSST(sst1, sst2) {
				t.Error("states shouldnot be equal")
			}

			if v1 && v2 {
				vvt1 := vv1.(BotMultiChoiceKeyboardControl)
				vvt2 := vv2.(BotMultiChoiceKeyboardControl)

				vst1 := SortedStringMap(vvt1.SelectedValues)
				vst2 := SortedStringMap(vvt2.SelectedValues)

				if IsEqualSST(vst1, vst2) {
					t.Error("values shouldnot be equal")
				}
			}

			if v1 {
				syncChan <- true
			}

			if v2 {
				syncChan <- true
			}

			if !v1 && !v2 {
				break
			}
		}
	}()

	wg.Wait()
	close(syncChan)
}
