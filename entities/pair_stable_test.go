package entities

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func assertNil(err error) {
	if err != nil {
		panic(err)
	}
}

func TestStablePair_GetOutputAmount(t *testing.T) {
	usdt := common.HexToAddress("0x493257fd37edb34451f62edf8d2a0c418852ba4c")
	usdc := common.HexToAddress("0x3355df6d4c9c3035724fd0e3914de96a5a83aaf4")
	tokenAmountA, err := NewTokenAmount(&Token{
		Address: usdc,
		Currency: &Currency{
			Decimals: 6,
			Symbol:   "USDC",
			Name:     "USDC",
		}}, big.NewInt(1372142240197))
	assertNil(err)
	tokenAmountB, err := NewTokenAmount(&Token{
		Address: usdt,
		Currency: &Currency{
			Decimals: 6,
			Symbol:   "USDT",
			Name:     "USDT",
		}}, big.NewInt(2953156372225))
	assertNil(err)
	tokenAmounts, err := NewTokenAmounts(tokenAmountA, tokenAmountB)
	assertNil(err)
	multiplier := big.NewInt(1e12)
	inAmount := big.NewInt(100000000) // 100usdt
	inputAmount, err := NewTokenAmount(&Token{
		Address: usdt,
		Currency: &Currency{
			Decimals: 6,
			Symbol:   "USDT",
			Name:     "USDT",
		}}, inAmount)
	assertNil(err)
	want, err := NewTokenAmount(&Token{
		Address: usdc,
		Currency: &Currency{
			Decimals: 6,
			Symbol:   "USDC",
			Name:     "USDC",
		}}, big.NewInt(99822875))
	assertNil(err)
	type fields struct {
		basePair    basePair
		multiplierA *big.Int
		multiplierB *big.Int
		fee         *big.Int
		feeBase     *big.Int
	}
	type args struct {
		inputAmount *TokenAmount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TokenAmount
		want1   Pair
		wantErr bool
	}{
		{
			name: "amount out",
			fields: fields{
				basePair:    basePair{TokenAmounts: tokenAmounts},
				multiplierA: multiplier,
				multiplierB: multiplier,
				fee:         big.NewInt(8),
				feeBase:     big.NewInt(10000),
			},
			args: args{
				inputAmount: inputAmount,
			},
			want:    want,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &StablePair{
				basePair:    tt.fields.basePair,
				multiplierA: tt.fields.multiplierA,
				multiplierB: tt.fields.multiplierB,
				fee:         tt.fields.fee,
				feeBase:     tt.fields.feeBase,
			}
			got, _, err := p.GetOutputAmount(tt.args.inputAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("StablePair.GetOutputAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StablePair.GetOutputAmount() got = %v, want %v", got, tt.want)
			}
		})
	}
}
