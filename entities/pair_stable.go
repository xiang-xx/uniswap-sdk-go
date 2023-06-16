package entities

import (
	"errors"
	"math/big"

	"github.com/xiang-xx/uniswap-sdk-go/constants"
	"github.com/xiang-xx/uniswap-sdk-go/utils"
)

// StablePair wrap stable pair
type StablePair struct {
	basePair

	multiplierA *big.Int
	multiplierB *big.Int
	fee         *big.Int
	feeBase     *big.Int
}

func NewStablePair(tokenAmountA, tokenAmountB *TokenAmount, multiplierA, multiplierB *big.Int) (Pair, error) {
	tokenAmounts, err := NewTokenAmounts(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, err
	}

	pair := &StablePair{
		basePair: basePair{
			TokenAmounts: tokenAmounts,
		},
		multiplierA: multiplierA,
		multiplierB: multiplierB,
		fee:         constants.Three,
		feeBase:     constants.B1000,
	}
	pair.LiquidityToken, err = NewToken(tokenAmountA.Token.ChainID, pair.GetAddress(),
		constants.Decimals18, constants.Univ2Symbol, constants.Univ2Name)
	return pair, err
}

func NewStablePairWithFee(tokenAmountA, tokenAmountB *TokenAmount, multiplierA, multiplierB *big.Int, fee uint64, feeBase uint64) (Pair, error) {
	tokenAmounts, err := NewTokenAmounts(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, err
	}

	pair := &StablePair{
		basePair: basePair{
			TokenAmounts: tokenAmounts,
		},
		multiplierA: multiplierA,
		multiplierB: multiplierB,
		fee:         big.NewInt(int64(fee)),
		feeBase:     big.NewInt(int64(feeBase)),
	}
	pair.LiquidityToken, err = NewToken(tokenAmountA.Token.ChainID, pair.GetAddress(),
		constants.Decimals18, constants.Univ2Symbol, constants.Univ2Name)
	return pair, err
}

/**** stable pair *****/

func (p *StablePair) PairType() PairType {
	return Stable
}

// GetOutputAmount returns OutputAmount and a Pair for the InputAmout
func (p *StablePair) GetOutputAmount(inputAmount *TokenAmount) (*TokenAmount, Pair, error) {
	if !p.InvolvesToken(inputAmount.Token) {
		return nil, nil, ErrDiffToken
	}

	if p.Reserve0().Raw().Cmp(constants.Zero) == 0 ||
		p.Reserve1().Raw().Cmp(constants.Zero) == 0 {
		return nil, nil, ErrInsufficientReserves
	}

	inputReserve, err := p.ReserveOf(inputAmount.Token)
	if err != nil {
		return nil, nil, err
	}
	token := p.Token0()
	if inputAmount.Token.Equals(p.Token0()) {
		token = p.Token1()
	}
	outputReserve, err := p.ReserveOf(token)
	if err != nil {
		return nil, nil, err
	}

	if utils.IsZero(inputAmount.CurrencyAmount.Numerator) {
		outputAmount, err := NewTokenAmount(token, big.NewInt(0))
		if err != nil {
			return nil, nil, err
		}
		return outputAmount, p, nil
	}

	_adjustedReserve0 := new(big.Int).Mul(p.Reserve0().Raw(), p.multiplierA)
	_adjustedReserve1 := new(big.Int).Mul(p.Reserve1().Raw(), p.multiplierB)

	_feeIn := new(big.Int).Div(new(big.Int).Mul(inputAmount.Raw(), p.fee), p.feeBase)
	_feeDeductedAmountIn := new(big.Int).Sub(inputAmount.Raw(), _feeIn)
	_d := utils.ComputeDFromAdjustedBalances(_adjustedReserve0, _adjustedReserve1)
	var outputAmount *big.Int
	if inputAmount.Token.Equals(p.Token0()) {
		_x := new(big.Int).Add(_adjustedReserve0, new(big.Int).Mul(_feeDeductedAmountIn, p.multiplierA))
		_y := utils.GetY(_x, _d)
		outputAmount = new(big.Int).Sub(
			new(big.Int).Sub(_adjustedReserve1, _y),
			constants.One,
		)
		outputAmount = new(big.Int).Div(outputAmount, p.multiplierB)
	} else {
		_x := new(big.Int).Add(_adjustedReserve1, new(big.Int).Mul(_feeDeductedAmountIn, p.multiplierB))
		_y := utils.GetY(_x, _d)
		outputAmount = new(big.Int).Sub(
			new(big.Int).Sub(_adjustedReserve0, _y),
			constants.One,
		)
		outputAmount = new(big.Int).Div(outputAmount, p.multiplierA)
	}

	if outputAmount.Cmp(constants.Zero) == 0 {
		return nil, nil, ErrInsufficientInputAmount
	}

	outputTokenAmount, err := NewTokenAmount(token, outputAmount)
	if err != nil {
		return nil, nil, err
	}

	tokenAmountA, err := inputAmount.Add(inputReserve)
	if err != nil {
		return nil, nil, err
	}
	tokenAmountB, err := outputReserve.Subtract(outputTokenAmount)
	if err != nil {
		return nil, nil, err
	}
	pair, err := NewPair(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, nil, err
	}
	return outputTokenAmount, pair, nil
}

// GetInputAmount returns InputAmout and a Pair for the OutputAmount
func (p *StablePair) GetInputAmount(outputAmount *TokenAmount) (*TokenAmount, Pair, error) {
	if !p.InvolvesToken(outputAmount.Token) {
		return nil, nil, ErrDiffToken
	}

	outputReserve, err := p.ReserveOf(outputAmount.Token)
	if err != nil {
		return nil, nil, err
	}
	if p.Reserve0().Raw().Cmp(constants.Zero) == 0 ||
		p.Reserve1().Raw().Cmp(constants.Zero) == 0 ||
		outputAmount.Raw().Cmp(outputReserve.Raw()) >= 0 {
		return nil, nil, ErrInsufficientReserves
	}

	token := p.Token0()
	if outputAmount.Token.Equals(p.Token0()) {
		token = p.Token1()
	}
	inputReserve, err := p.ReserveOf(token)
	if err != nil {
		return nil, nil, err
	}

	_adjustedReserve0 := new(big.Int).Mul(p.Reserve0().Raw(), p.multiplierA)
	_adjustedReserve1 := new(big.Int).Mul(p.Reserve1().Raw(), p.multiplierB)
	_d := utils.ComputeDFromAdjustedBalances(_adjustedReserve0, _adjustedReserve1)

	var inputAmount *big.Int
	if outputAmount.Token.Equals(p.Token0()) {
		_y := new(big.Int).Sub(
			_adjustedReserve0,
			new(big.Int).Mul(outputAmount.Raw(), p.multiplierA),
		)
		if _y.Cmp(constants.One) <= 0 {
			inputAmount = constants.One
		} else {
			_x := utils.GetY(_y, _d)
			inputAmount = new(big.Int).Add(constants.One,
				new(big.Int).Div(
					new(big.Int).Mul(p.feeBase, new(big.Int).Sub(_x, _adjustedReserve1)),
					new(big.Int).Sub(p.feeBase, p.fee),
				),
			)
			inputAmount = new(big.Int).Div(inputAmount, p.multiplierB)
		}
	} else {
		_y := new(big.Int).Sub(
			_adjustedReserve1,
			new(big.Int).Mul(outputAmount.Raw(), p.multiplierB),
		)
		if _y.Cmp(constants.One) <= 0 {
			inputAmount = constants.One
		} else {
			_x := utils.GetY(_y, _d)
			inputAmount = new(big.Int).Add(constants.One,
				new(big.Int).Div(
					new(big.Int).Mul(p.feeBase, new(big.Int).Sub(_x, _adjustedReserve0)),
					new(big.Int).Sub(p.feeBase, p.fee),
				),
			)
			inputAmount = new(big.Int).Div(inputAmount, p.multiplierA)
		}
	}

	inputTokenAmount, err := NewTokenAmount(token, inputAmount)
	if err != nil {
		return nil, nil, err
	}

	tokenAmountA, err := inputTokenAmount.Add(inputReserve)
	if err != nil {
		return nil, nil, err
	}
	tokenAmountB, err := outputReserve.Subtract(outputAmount)
	if err != nil {
		return nil, nil, err
	}
	pair, err := NewPair(tokenAmountA, tokenAmountB)
	if err != nil {
		return nil, nil, err
	}
	return inputTokenAmount, pair, nil
}

// GetLiquidityMinted returns liquidity minted TokenAmount
func (p *StablePair) GetLiquidityMinted(totalSupply, tokenAmountA, tokenAmountB *TokenAmount) (*TokenAmount, error) {
	return nil, errors.New("StablePair not implement totalSupply")
}

// GetLiquidityValue returns liquidity value TokenAmount
func (p *StablePair) GetLiquidityValue(token *Token, totalSupply, liquidity *TokenAmount, feeOn bool, kLast *big.Int) (*TokenAmount, error) {
	return nil, errors.New("StablePair not implement GetLiquidityValue")
}
