package helpers

import (
	"math/big"
)

func IsValidGroupID(id string) bool {

	if len(id) > 8 && len(id) != 18 {
		return false
	}

	i := big.NewInt(0)

	// Help GC
	defer func(i *big.Int) {
		i = nil
	}(i)

	i, success := i.SetString(id, 10)
	if !success || i == nil {
		return false
	}

	return true
}
