package botform

import "testing"

func TestSortedStringMap_Compare(t *testing.T) {
	ssm1 := SortedStringMap{"a": "1", "b": "2"}
	ssm2 := SortedStringMap{"b": "2", "a": "1"}

	if !IsEqualSST(ssm1, ssm2) {
		t.Error("ssm1 is not equal ssm2")
	}

	ssm3 := SortedStringMap{"b": "2", "a": "2"}

	if IsEqualSST(ssm1, ssm3) {
		t.Error("ssm1 is equal ssm2")
	}
}
