package botform

import (
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func ReverseSlice[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

type SortedStringMap map[string]string

func (ssm SortedStringMap) Range(cb func(k, v string) bool, byValue bool) {
	keys := make([]string, 0, len(ssm))
	valueIndex := map[string]string{}

	if byValue {
		for k, v := range ssm {
			keys = append(keys, v)
			valueIndex[v] = k
		}
	} else {
		for k := range ssm {
			keys = append(keys, k)
		}
	}
	c := collate.New(language.Und, collate.IgnoreCase)
	c.SortStrings(keys)
	for _, mapKey := range keys {
		if byValue {
			k := valueIndex[mapKey]
			ok := cb(k, mapKey)
			if !ok {
				break
			}
		} else {
			ok := cb(mapKey, ssm[mapKey])
			if !ok {
				break
			}
		}
	}
}

func IsEqualSST(ssm1 SortedStringMap, ssm2 SortedStringMap) bool {

	s1 := ""
	ssm1.Range(func(k, v string) bool {
		s1 = s1 + v
		return true
	}, false)

	s2 := ""
	ssm2.Range(func(k, v string) bool {
		s2 = s2 + v
		return true
	}, false)

	if s1 == s2 {
		return true
	}

	return false
}

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
