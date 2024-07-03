package resp

import "bytes"

type RESPSimpleString struct {
	Value string
}

func (s *RESPSimpleString) Type() string {
	return "SimpleString"
}

func (s *RESPSimpleString) String() string {
	return s.Value
}

func (s *RESPSimpleString) Equal(other RESPValue) bool {
	otherString, ok := other.(*RESPSimpleString)
	if !ok {
		return false
	}
	return s.Value == otherString.Value
}

func (ss *RESPSimpleString) Encode(buf *bytes.Buffer) error {
	buf.WriteByte('+')
	buf.WriteString(ss.Value)
	buf.WriteString("\r\n")
	return nil
}
