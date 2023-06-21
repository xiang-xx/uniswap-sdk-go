package entities

type SmartTrade struct {
	// len(Percents) == len(Trades)
	Percents []int
	Trades   []*Trade

	inputAmount  *TokenAmount
	outputAmount *TokenAmount
}

func (t *SmartTrade) InputAmount() *TokenAmount {
	return t.inputAmount
}

func (t *SmartTrade) OutputAmount() *TokenAmount {
	return t.outputAmount
}

type tradesWithPercent struct {
	Percents      []int
	Trades        []*Trade
	CurrentPairs  []Pair
	RemainPercent int
	PercentIndex  int
}
