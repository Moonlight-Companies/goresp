package resp

import (
	"bytes"
	"strconv"
)

type RESPBulkString struct {
	Value []byte
}

func (b *RESPBulkString) Type() string {
	return "BulkString"
}

func (b *RESPBulkString) String() string {
	if b.Value == nil {
		return "<nil>"
	}
	return string(b.Value)
}

func (s *RESPBulkString) Equal(other RESPValue) bool {
	otherString, ok := other.(*RESPBulkString)
	if !ok {
		return false
	}

	return string(s.Value) == string(otherString.Value)
}

func (bs *RESPBulkString) Encode(buf *bytes.Buffer) error {
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(len(bs.Value)))
	buf.WriteString("\r\n")
	buf.Write(bs.Value)
	buf.WriteString("\r\n")
	return nil
}
