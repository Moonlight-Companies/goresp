package resp

import (
	"bytes"
	"errors"
)

var errIncompleteData = errors.New("incomplete data")
var errUnrecoverableProtocol = errors.New("unrecoverable protocol error")

type RESPValue interface {
	Type() string
	String() string
	Equal(RESPValue) bool
	Decode(*bytes.Buffer, int) (int, error)
	Encode(buf *bytes.Buffer) error
}

func DecodeValue(buf *bytes.Buffer, start int) (RESPValue, int, error) {
	if buf.Len() == 0 {
		return nil, 0, nil
	}

	if start >= buf.Len() {
		return nil, 0, errIncompleteData
	}

	switch buf.Bytes()[start] {
	case byte(SIMPLE_STRING):
		s := &RESPSimpleString{}
		n, err := s.Decode(buf, start)
		return s, n, err
	case byte(INTEGER):
		i := &RESPInteger{}
		n, err := i.Decode(buf, start)
		return i, n, err
	case byte(BULK_STRING):
		b := &RESPBulkString{}
		n, err := b.Decode(buf, start)
		return b, n, err
	case byte(ERROR):
		e := &RESPError{}
		n, err := e.Decode(buf, start)
		return e, n, err
	case byte(ARRAY):
		e := &RESPArray{}
		n, err := e.Decode(buf, start)
		return e, n, err
	default:
		return nil, 0, errors.New("invalid opcode")
	}
}
