package utils

import (
	"math/big"

	"github.com/xiang-xx/uniswap-sdk-go/constants"
)

func GetY(x, d *big.Int) *big.Int {
	// c = (d * d) / (x * 2);
	c := mulDiv(d, d, mul(constants.Two, x))
	//c = (c * d) / 4000;
	c = mulDiv(c, d, constants.B4000)

	// b = x + (d / 2000)
	b := new(big.Int).Add(x, new(big.Int).Div(d, constants.B2000))
	yPrev := new(big.Int)
	y := d

	/// @dev Iterative approximation.
	for i := 0; i < 256; {
		yPrev = y
		//y = (y * y + c) / (y * 2 + b - d);
		y = new(big.Int).Div(
			new(big.Int).Add(
				new(big.Int).Mul(y, y),
				c,
			),
			new(big.Int).Sub(
				new(big.Int).Add(
					mul(constants.Two, y),
					b,
				),
				d,
			),
		)

		if within1(y, yPrev) {
			break
		}

		i++
	}
	return y
}

// Overflow checks should be applied before calling this function.
// The maximum XPs are `3802571709128108338056982581425910818` of uint128.
func ComputeDFromAdjustedBalances(xp0, xp1 *big.Int) *big.Int {
	s := add(xp0, xp1)

	computed := new(big.Int)
	if IsZero(s) {
		computed = constants.Zero
	} else {
		prevD := new(big.Int)
		d := s

		for i := 0; i < 256; {
			//uint dP = (((d * d) / xp0) * d) / xp1 / 4;
			dP := div(mulDiv(mulDiv(d, d, xp0), d, xp1), constants.Four)

			prevD = d
			//d = (((2000 * s) + 2 * dP) * d) / ((2000 - 1) * d + 3 * dP);
			d = mulDiv(
				// `s` cannot be zero and this value will never be zero.
				add(mul(constants.B2000, s), mul(constants.Two, dP)),
				d,
				add(mul(constants.B1999, d), mul(constants.Three, dP)),
			)

			if within1(d, prevD) {
				break
			}

			i++
		}

		computed = d
	}
	return computed
}
