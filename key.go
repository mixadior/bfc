package botform

import "github.com/vmihailenco/msgpack/v5"

type BfcKey []string

func (b BfcKey) GetLast() string {
	if len(b) == 0 {
		return ""
	}
	return b[len(b)-1]
}

func (b BfcKey) Shift() BfcKey {
	bb := b[0 : len(b)-1]
	return bb
}

func (b BfcKey) Marshal() string {
	mv, _ := msgpack.Marshal(b)
	return string(mv)
}

func UnMarshalBfcKey(mkey string) (b BfcKey, e error) {
	b = BfcKey{}
	e = msgpack.Unmarshal([]byte(mkey), &b)
	return b, e
}
