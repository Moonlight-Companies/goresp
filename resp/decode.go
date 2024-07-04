package resp

import (
	"bytes"
)

type Decode struct {
	buffer bytes.Buffer
}

func NewDecode() *Decode {
	return &Decode{}
}

// Provide adds data to the parser's buffer
func (p *Decode) Provide(data []byte) {
	p.buffer.Write(data)
}

// Reset clears the parser's buffer, usually after a reconnect
func (p *Decode) Reset() {
	p.buffer.Reset()
}

// Parse attempts to parse a complete RESP value from the current buffer
func (p *Decode) Parse() (RESPValue, error) {
	value, bytesConsumed, err := DecodeValue(&p.buffer, 0)
	if err != nil {
		if err == errIncompleteData {
			return nil, nil
		}
		return nil, err
	}

	p.buffer.Next(bytesConsumed)
	return value, nil
}

// HasData checks if there's enough data in the buffer to potentially parse a complete RESP value
// its 3 because we must have a Opcode and a CRLF at the minimum with something between them
func (p *Decode) HasData() bool {
	return p.buffer.Len() > 3
}
