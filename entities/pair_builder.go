package entities

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xiang-xx/uniswap-sdk-go/constants"
)

type PairBuilder struct {
	tokenAmountA *TokenAmount
	tokenAmountB *TokenAmount
	fee          uint64
	feeBase      uint64
	pairAddress  common.Address
	// if multipliers is all nil or is all 1, pair is classic pair
	// else, is stable pair
	multiplierA *big.Int
	multiplierB *big.Int
}

func NewPairBuilder() *PairBuilder {
	return &PairBuilder{}
}

func (p *PairBuilder) SetTokenAmountA(tokenAmountA *TokenAmount) *PairBuilder {
	p.tokenAmountA = tokenAmountA
	return p
}

func (p *PairBuilder) SetTokenAmountB(tokenAmountB *TokenAmount) *PairBuilder {
	p.tokenAmountB = tokenAmountB
	return p
}

func (p *PairBuilder) SetTokenAmounts(tokenAmountA, tokenAmountB *TokenAmount) *PairBuilder {
	p.tokenAmountA = tokenAmountA
	p.tokenAmountB = tokenAmountB
	return p
}

func (p *PairBuilder) SetFee(fee, feeBase uint64) *PairBuilder {
	p.fee = fee
	p.feeBase = feeBase
	return p
}

func (p *PairBuilder) SetPairAddress(pairAddress common.Address) *PairBuilder {
	p.pairAddress = pairAddress
	return p
}

// SetTokenMultiplier set pair as table pair
func (p *PairBuilder) SetTokenMultiplier(multiplierA, multiplierB *big.Int) *PairBuilder {
	p.multiplierA = multiplierA
	p.multiplierB = multiplierB
	return p
}

func (p *PairBuilder) Build() (Pair, error) {
	if nil == p.tokenAmountA || nil == p.tokenAmountB {
		return nil, errors.New("token amount not set")
	}
	tokenAmounts, err := NewTokenAmounts(p.tokenAmountA, p.tokenAmountB)
	if err != nil {
		return nil, err
	}

	// set default fee 3/1000
	fee := p.fee
	feeBase := p.feeBase
	var (
		feeBI     *big.Int
		feeBaseBI *big.Int
	)
	if fee == 0 && feeBase == 0 {
		feeBI = constants.Three
		feeBaseBI = constants.B1000
	} else {
		feeBI = big.NewInt(int64(fee))
		feeBaseBI = big.NewInt(int64(feeBase))
	}

	if p.multiplierA == nil || p.multiplierB == nil ||
		(p.multiplierA.Uint64() <= 1 && p.multiplierB.Uint64() <= 1) {
		pair := &ClassicPair{
			basePair: basePair{
				TokenAmounts: tokenAmounts,
				PairAddress:  p.pairAddress,
			},
			fee:     feeBI,
			feeBase: feeBaseBI,
		}
		pair.LiquidityToken, err = NewToken(p.tokenAmountA.Token.ChainID, pair.GetAddress(),
			constants.Decimals18, constants.Univ2Symbol, constants.Univ2Name)
		return pair, err
	} else {
		pair := &StablePair{
			basePair: basePair{
				TokenAmounts: tokenAmounts,
				PairAddress:  p.pairAddress,
			},
			multiplierA: p.multiplierA,
			multiplierB: p.multiplierB,
			fee:         feeBI,
			feeBase:     feeBaseBI,
		}
		pair.LiquidityToken, err = NewToken(p.tokenAmountA.Token.ChainID, pair.GetAddress(),
			constants.Decimals18, constants.Univ2Symbol, constants.Univ2Name)
		return pair, err
	}
}
