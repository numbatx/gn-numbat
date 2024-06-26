package state

// address is the struct holding an Numbat address identifier
type address struct {
	bytes []byte
}

// NewAddress creates a new Address with the same byte slice as the parameter received
func NewAddress(adr []byte) *address {
	return &address{bytes: adr}
}

// Bytes returns the data corresponding to this address
func (adr *address) Bytes() []byte {
	return adr.bytes
}
