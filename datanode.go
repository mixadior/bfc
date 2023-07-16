package botform

type BotFormDataNode struct {
	Children map[string]*BotFormDataNode
	Value    string
	Name     string
	Hidden   bool
	Meta     map[string]interface{}
}
