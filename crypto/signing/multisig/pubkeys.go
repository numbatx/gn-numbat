package multisig

import (
	"github.com/numbatx/gn-numbat/crypto"
)

func convertStringsToPubKeys(pubKeys []string, kg crypto.KeyGenerator) ([]crypto.PublicKey, error) {
	var pk []crypto.PublicKey

	//convert pubKeys
	for _, pubKeyStr := range pubKeys {
		if pubKeyStr == "" {
			return nil, crypto.ErrEmptyPubKeyString
		}

		pubKey, err := kg.PublicKeyFromByteArray([]byte(pubKeyStr))
		if err != nil {
			return nil, crypto.ErrInvalidPublicKeyString
		}

		pk = append(pk, pubKey)
	}
	return pk, nil
}
