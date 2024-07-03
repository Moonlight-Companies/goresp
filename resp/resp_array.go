package resp

import "fmt"

type RESPArray struct {
	Items []RESPValue
}

func (a *RESPArray) Type() string {
	return "Array"
}

func (a *RESPArray) String() string {
	return fmt.Sprintf("%v", a.Items)
}

func (s *RESPArray) Equal(other RESPValue) bool {
	otherString, ok := other.(*RESPArray)
	if !ok {
		return false
	}
	if len(s.Items) != len(otherString.Items) {
		return false
	}

	for i, item := range s.Items {
		if !item.Equal(otherString.Items[i]) {
			return false
		}
	}

	return true
}
