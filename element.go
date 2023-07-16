package botform

import (
	"errors"
	"gopkg.in/yaml.v2"
)

type BotFormElementNode struct {
	Renderer    BotControl `yaml:"-"`
	RendererRaw []byte     `yaml:"renderer"`
	RenderTitle string
	DataPath    []string
	Value       string
	Required    bool
	Nested      bool
}

func (bfen *BotFormElementNode) MarshalYAML() (interface{}, error) {
	if bfen.Renderer != nil {
		out, err := yaml.Marshal(bfen.Renderer)
		if err != nil {
			return nil, err
		}
		bfen.RendererRaw = out
		bfen.RenderTitle = bfen.Renderer.GetTitle()
	}
	return bfen, nil
}

func (bfen *BotFormElementNode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawBotFormElementNode BotFormElementNode
	if err := unmarshal((*rawBotFormElementNode)(bfen)); err != nil {
		return err
	}

	rMap := map[string]string{}
	yaml.Unmarshal(bfen.RendererRaw, &rMap)

	rendererType, ok := rMap["type"]
	if !ok {
		return errors.New("renderer has no type field.")
	}

	switch rendererType {
	case "BotChoiceKeyboardControl":
		v := BotChoiceKeyboardControl{}
		err := yaml.Unmarshal(bfen.RendererRaw, &v)
		if err != nil {
			return err
		}
		bfen.Renderer = v
	case "BotMultiChoiceKeyboardControl":
		v := BotMultiChoiceKeyboardControl{}
		err := yaml.Unmarshal(bfen.RendererRaw, &v)
		if err != nil {
			return err
		}
		bfen.Renderer = v
	case "BotTextKeyboardControl":
		v := BotTextKeyboardControl{}
		err := yaml.Unmarshal(bfen.RendererRaw, &v)
		if err != nil {
			return err
		}
		bfen.Renderer = v
	}

	return nil
}

type BotControl interface {
	GetTitle() string
}
