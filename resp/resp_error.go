package resp

type RESPError struct {
	Value string
}

func (e *RESPError) Type() string {
	return "Error"
}

func (e *RESPError) String() string {
	return e.Value
}

func (s *RESPError) Equal(other RESPValue) bool {
	otherError, ok := other.(*RESPError)
	if !ok {
		return false
	}
	return s.Value == otherError.Value
}
