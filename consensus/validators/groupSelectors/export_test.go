package groupSelectors

import (
	"github.com/numbatx/gn-numbat/consensus"
)

func (ihgs *indexHashedGroupSelector) EligibleList() []consensus.Validator {
	return ihgs.eligibleList
}
