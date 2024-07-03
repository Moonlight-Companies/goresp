package resp

import "fmt"

type RESPInteger struct {
	Value int64
}

func (i *RESPInteger) Type() string {
	return "Integer"
}

func (i *RESPInteger) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (s *RESPInteger) Equal(other RESPValue) bool {
	otherInteger, ok := other.(*RESPInteger)
	if !ok {
		return false
	}
	return s.Value == otherInteger.Value
}
