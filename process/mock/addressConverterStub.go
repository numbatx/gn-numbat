package mock

import (
	"github.com/numbatx/gn-numbat/data/state"
)

type AddressConverterStub struct {
	CreateAddressFromPublicKeyBytesCalled func(pubKey []byte) (state.AddressContainer, error)
	ConvertToHexCalled                    func(addressContainer state.AddressContainer) (string, error)
	CreateAddressFromHexCalled            func(hexAddress string) (state.AddressContainer, error)
	PrepareAddressBytesCalled             func(addressBytes []byte) ([]byte, error)
}

func (acs *AddressConverterStub) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	return acs.CreateAddressFromPublicKeyBytesCalled(pubKey)
}

func (acs *AddressConverterStub) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	return acs.ConvertToHexCalled(addressContainer)
}

func (acs *AddressConverterStub) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	return acs.CreateAddressFromHexCalled(hexAddress)
}

func (acs *AddressConverterStub) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	return acs.PrepareAddressBytesCalled(addressBytes)
}
