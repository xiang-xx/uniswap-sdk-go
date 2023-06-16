package entities

import (
	"math/big"
	"reflect"
	"testing"
)

func TestStablePair_GetOutputAmount(t *testing.T) {
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
		// TODO: Add test cases.
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
			got, got1, err := p.GetOutputAmount(tt.args.inputAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("StablePair.GetOutputAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StablePair.GetOutputAmount() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("StablePair.GetOutputAmount() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
