package entities

type SmartTrade struct {
	// len(Percents) == len(Trades)
	Percents []int
	Trades   []*Trade

	OutputAmount *TokenAmount
}

type tradesWithPercent struct {
	Percents      []int
	Trades        []*Trade
	CurrentPairs  []Pair
	RemainPercent int
	PercentIndex  int
}
