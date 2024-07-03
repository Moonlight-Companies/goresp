package resp

type RESPValue interface {
	Type() string
	String() string
	Equal(RESPValue) bool
}
