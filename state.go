package botform

import (
	"github.com/vmihailenco/msgpack/v5"
)

type State struct {
	Values                 map[string]string
	Key                    BfcKey
	CancelActionLabel      string
	SaveAndNextActionLabel string
	NextActionLabel        string
	PrevActionLabel        string
	IsFinished             bool
	IsCanceled             bool
}

func (s *State) GetValues() map[string]string {
	return s.Values
}

func (s *State) Store(k, v string) {
	s.Values[k] = v
}

func (s *State) Delete(k string) {
	delete(s.Values, k)
}

func (s *State) Marshal() (string, error) {
	m, e := msgpack.Marshal(s)
	return string(m), e
}

func NewState(data map[string]string, key BfcKey) *State {
	ss := &State{Values: data, Key: key, IsCanceled: false, IsFinished: false, CancelActionLabel: "[X]", PrevActionLabel: "<=", SaveAndNextActionLabel: "Save and next", NextActionLabel: "Skip step"}
	return ss
}

func UnmarshalState(data []byte) (s *State, e error) {
	s = &State{}
	e = msgpack.Unmarshal(data, &s)
	if e != nil {
		return nil, e
	}
	return s, nil
}
