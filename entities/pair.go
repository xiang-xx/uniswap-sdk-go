package entities

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/xiang-xx/uniswap-sdk-go/constants"
)

type PairType string

var (
	_PairAddressCache = &PairAddressCache{
		lk:      new(sync.RWMutex),
		address: make(map[common.Address]map[common.Address]common.Address, 16),
	}

	// ErrInvalidLiquidity invalid liquidity
	ErrInvalidLiquidity = fmt.Errorf("invalid liquidity")
	// ErrInvalidKLast invalid kLast
	ErrInvalidKLast = fmt.Errorf("invalid kLast")

	Classic PairType = "classic"
	Stable  PairType = "stable"
)

// TokenAmounts warps TokenAmount array
type TokenAmounts [2]*TokenAmount

// Tokens warps Token array
type Tokens [2]*Token

// NewTokenAmounts creates a TokenAmount
func NewTokenAmounts(tokenAmountA, tokenAmountB *TokenAmount) (TokenAmounts, error) {
	ok, err := tokenAmountA.Token.SortsBefore(tokenAmountB.Token)
	if err != nil {
		return TokenAmounts{}, err
	}
	if ok {
		return TokenAmounts{tokenAmountA, tokenAmountB}, nil
	}
	return TokenAmounts{tokenAmountB, tokenAmountA}, nil
}

// PairAddressCache warps pair address cache
type PairAddressCache struct {
	lk *sync.RWMutex
	// token0 address : token1 address : pair address
	address map[common.Address]map[common.Address]common.Address
}

// GetAddress returns contract address
// addressA < addressB
func (p *PairAddressCache) GetAddress(addressA, addressB common.Address) common.Address {
	p.lk.RLock()
	pairAddresses, ok := p.address[addressA]
	if !ok {
		p.lk.RUnlock()
		p.lk.Lock()
		defer p.lk.Unlock()
		addr := getCreate2Address(addressA, addressB)
		p.address[addressA] = map[common.Address]common.Address{
			addressB: addr,
		}
		return addr
	}

	pairAddress, ok := pairAddresses[addressB]
	if !ok {
		p.lk.RUnlock()
		p.lk.Lock()
		defer p.lk.Unlock()
		addr := getCreate2Address(addressA, addressB)
		pairAddresses[addressB] = addr
		return addr
	}

	p.lk.RUnlock()
	return pairAddress
}

func getCreate2Address(addressA, addressB common.Address) common.Address {
	var salt [32]byte
	copy(salt[:], crypto.Keccak256(append(addressA.Bytes(), addressB.Bytes()...)))
	return crypto.CreateAddress2(constants.FactoryAddress, salt, constants.InitCodeHash)
}

type Pair interface {
	ChainID() constants.ChainID
	GetAddress() common.Address
	GetInputAmount(outputAmount *TokenAmount) (*TokenAmount, Pair, error)
	GetLiquidityMinted(totalSupply *TokenAmount, tokenAmountA *TokenAmount, tokenAmountB *TokenAmount) (*TokenAmount, error)
	GetLiquidityValue(token *Token, totalSupply *TokenAmount, liquidity *TokenAmount, feeOn bool, kLast *big.Int) (*TokenAmount, error)
	GetOutputAmount(inputAmount *TokenAmount) (*TokenAmount, Pair, error)
	InvolvesToken(token *Token) bool
	PriceOf(token *Token) (*Price, error)
	Reserve0() *TokenAmount
	Reserve1() *TokenAmount
	ReserveOf(token *Token) (*TokenAmount, error)
	Token0() *Token
	Token0Price() *Price
	Token1() *Token
	Token1Price() *Price
	PairType() PairType
	GetLiquidityToken() *Token
}
