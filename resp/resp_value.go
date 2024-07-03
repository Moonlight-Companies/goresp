package resp

import "bytes"

type RESPValue interface {
	Type() string
	String() string
	Equal(RESPValue) bool
	Encode(buf *bytes.Buffer) error
}
