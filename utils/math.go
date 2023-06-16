package utils

import (
	"math/big"

	"github.com/xiang-xx/uniswap-sdk-go/constants"
)

func within1(a, b *big.Int) bool {
	if a.Cmp(b) > 0 {
		// a - b <= 1
		return new(big.Int).Sub(a, b).Cmp(constants.One) <= 0
	}
	// b - a <= 1;
	return new(big.Int).Sub(b, a).Cmp(constants.One) <= 0
}

func IsZero(x *big.Int) bool {
	return x.Cmp(constants.Zero) == 0
}

func mulDiv(x, y, denominator *big.Int) *big.Int {
	return div(mul(x, y), denominator)
}

func add(x, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}

func mul(x, y *big.Int) *big.Int {
	return new(big.Int).Mul(x, y)
}

func div(x, y *big.Int) *big.Int {
	return new(big.Int).Div(x, y)
}
