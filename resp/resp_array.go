package resp

import (
	"bytes"
	"fmt"
	"strconv"
)

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

func (a *RESPArray) Encode(buf *bytes.Buffer) error {
	buf.WriteByte(byte(ARRAY))
	if a.Items == nil {
		buf.WriteString("-1")
		buf.Write(PROTOCOL_SEPARATOR)
		return nil
	}
	buf.WriteString(strconv.Itoa(len(a.Items)))
	buf.Write(PROTOCOL_SEPARATOR)
	for _, item := range a.Items {
		if err := item.Encode(buf); err != nil {
			return err
		}
	}
	return nil
}
