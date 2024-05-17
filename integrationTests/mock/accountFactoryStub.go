package mock

import "github.com/numbatx/gn-numbat/data/state"

type AccountsFactoryStub struct {
	CreateAccountCalled func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error)
}

func (afs *AccountsFactoryStub) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	return afs.CreateAccountCalled(address, tracker)
}
