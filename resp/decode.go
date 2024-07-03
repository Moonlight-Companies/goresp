package resp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
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
	value, bytesConsumed, err := p.parseValue(0)
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

var errIncompleteData = errors.New("incomplete data")
var errUnrecoverableProtocol = errors.New("unrecoverable protocol error")

func (p *Decode) parseValue(offset int) (RESPValue, int, error) {
	if offset >= p.buffer.Len() {
		return nil, 0, errIncompleteData
	}

	opcode := OPCODE(p.buffer.Bytes()[offset])
	switch opcode {
	case SIMPLE_STRING:
		return p.parseSimpleString(offset + 1)
	case ERROR:
		return p.parseError(offset + 1)
	case BULK_STRING:
		return p.parseBulkString(offset + 1)
	case INTEGER:
		return p.parseInteger(offset + 1)
	case ARRAY:
		return p.parseArray(offset + 1)
	default:
		return nil, 0, fmt.Errorf("%w: unsupported opcode %c at offset %d", errUnrecoverableProtocol, opcode, offset)
	}
}

func (p *Decode) findProtocolSeparator(offset int) (int, error) {
	index := bytes.Index(p.buffer.Bytes()[offset:], PROTOCOL_SEPARATOR)
	if index == -1 {
		return 0, errIncompleteData
	}
	return offset + index, nil
}

func (p *Decode) parseSimpleString(offset int) (RESPValue, int, error) {
	endIndex, err := p.findProtocolSeparator(offset)
	if err != nil {
		return nil, 0, err
	}
	value := string(p.buffer.Bytes()[offset:endIndex])
	return &RESPSimpleString{Value: value}, endIndex + 2, nil
}

func (p *Decode) parseError(offset int) (RESPValue, int, error) {
	endIndex, err := p.findProtocolSeparator(offset)
	if err != nil {
		return nil, 0, err
	}
	value := string(p.buffer.Bytes()[offset:endIndex])
	return &RESPError{Value: value}, endIndex + 2, nil
}

func (p *Decode) parseBulkString(offset int) (RESPValue, int, error) {
	length, lengthEndOffset, err := p.parseLength(offset)
	if err != nil {
		return nil, 0, err
	}
	if length == -1 {
		return &RESPBulkString{Value: nil}, lengthEndOffset, nil
	}

	endIndex := lengthEndOffset + length
	if endIndex+2 > p.buffer.Len() {
		return nil, 0, errIncompleteData
	}

	value := make([]byte, length)
	copy(value, p.buffer.Bytes()[lengthEndOffset:endIndex])
	return &RESPBulkString{Value: value}, endIndex + 2, nil
}

func (p *Decode) parseInteger(offset int) (RESPValue, int, error) {
	endIndex, err := p.findProtocolSeparator(offset)
	if err != nil {
		return nil, 0, err
	}
	value, err := strconv.ParseInt(string(p.buffer.Bytes()[offset:endIndex]), 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: invalid integer", errUnrecoverableProtocol)
	}
	return &RESPInteger{Value: value}, endIndex + 2, nil
}

func (p *Decode) parseArray(offset int) (RESPValue, int, error) {
	length, lengthEndOffset, err := p.parseLength(offset)
	if err != nil {
		return nil, 0, err
	}
	if length == -1 {
		return &RESPArray{Items: nil}, lengthEndOffset, nil
	}

	array := &RESPArray{Items: make([]RESPValue, 0, length)}
	currentOffset := lengthEndOffset

	for i := 0; i < length; i++ {
		item, newOffset, err := p.parseValue(currentOffset)
		if err != nil {
			return nil, 0, err
		}
		array.Items = append(array.Items, item)
		currentOffset = newOffset
	}

	return array, currentOffset, nil
}

func (p *Decode) parseLength(offset int) (int, int, error) {
	endIndex, err := p.findProtocolSeparator(offset)
	if err != nil {
		return 0, 0, err
	}
	length, err := strconv.Atoi(string(p.buffer.Bytes()[offset:endIndex]))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: invalid length", errUnrecoverableProtocol)
	}
	return length, endIndex + 2, nil
}
