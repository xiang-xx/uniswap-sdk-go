package entities

import (
	"math/big"

	"github.com/xiang-xx/uniswap-sdk-go/constants"
)

type BestSmartTradeOptions struct {
	BestTradeOptions

	// eg. the min unit split from inAmount
	// 1-100，eg 2/4/5/10/20, default is 10
	SplitPercentage int

	// maximum count to split inAmount
	MaxSplit int

	// how many results to return
	MaxSmartTradeNumResults int
}

func BestSmartTradeExactIn(
	pairs []Pair,
	currencyAmountIn *TokenAmount,
	currencyOut *Token,
	options *BestSmartTradeOptions) ([]*SmartTrade, error) {
	if nil == options {
		return nil, ErrInvalidOption
	}
	if len(pairs) == 0 {
		return nil, ErrInvalidPairs
	}
	splitPercentage := 10
	if options.SplitPercentage > 0 {
		if splitPercentage < 0 || splitPercentage > 100 {
			return nil, ErrInvalidOption
		}
		if (100/options.SplitPercentage)*options.SplitPercentage != 100 {
			return nil, ErrInvalidOption
		}
	}

	percents, amounts, err := getAmountDistribution(currencyAmountIn, splitPercentage)
	if err != nil {
		return nil, err
	}

	percentToTrades, err := getPercentageToTrades(percents, amounts, pairs, currencyOut, options)
	if err != nil {
		return nil, err
	}

	// BFS to get SmartTrades
	smartTrades := make([]*SmartTrade, 0)
	queue := make([]tradesWithPercent, 0)
	for i, percent := range percents {
		trades, ok := percentToTrades[percent]
		if !ok || len(trades) == 0 {
			continue
		}

		queue = append(queue, tradesWithPercent{
			RemainPercent: 100 - percent,
			PercentIndex:  i,
			Percents:      []int{percent},
			Trades:        []*Trade{trades[0]},
			CurrentPairs:  trades[0].Route.Pairs,
		})
	}

	split := 0
	for split < options.MaxSplit {
		split++

		nextQueue := make([]tradesWithPercent, 0)
		for _, item := range queue {
			for i := item.PercentIndex; i >= 0; i-- {
				percentA := percents[i]
				if percentA > item.RemainPercent {
					continue
				}
				trades, ok := percentToTrades[percentA]
				if !ok {
					continue
				}
				matchedTrade := findFirstTradeNotUsingPair(item.CurrentPairs, trades)
				if matchedTrade == nil {
					continue
				}

				remainPerent := item.RemainPercent - percentA
				currentPairs := append(item.CurrentPairs, matchedTrade.Route.Pairs...)
				currentPercents := append(item.Percents, percentA)
				currentTrades := append(item.Trades, matchedTrade)

				if remainPerent == 0 {
					outputAmount := currentTrades[0].OutputAmount()
					inputAmount := currentTrades[0].InputAmount()
					for k := 1; k < len(currentTrades); k++ {
						outputAmount, err = outputAmount.Add(currentTrades[k].OutputAmount())
						if err != nil {
							return nil, err
						}
						inputAmount, err = inputAmount.Add(currentTrades[k].InputAmount())
						if err != nil {
							return nil, err
						}
					}
					smartTrade := &SmartTrade{
						Percents:     currentPercents,
						Trades:       currentTrades,
						outputAmount: outputAmount,
						inputAmount:  inputAmount,
					}
					smartTrades, _, err = SortedInsert(smartTrades, smartTrade, options.MaxSmartTradeNumResults, SmartTradeComparator)
					if err != nil {
						return nil, err
					}
				} else {
					nextQueue = append(nextQueue, tradesWithPercent{
						Percents:      currentPercents,
						Trades:        currentTrades,
						CurrentPairs:  currentPairs,
						RemainPercent: remainPerent,
						PercentIndex:  i,
					})
				}
			}
		}
		queue = nextQueue
	}

	return smartTrades, nil
}

func findFirstTradeNotUsingPair(currentPairs []Pair, trades []*Trade) *Trade {
	for i, trade := range trades {
		existPair := false
	outer:
		for _, tradePair := range trade.Route.Pairs {
			for _, pair := range currentPairs {
				if tradePair.Equal(pair) {
					existPair = true
					break outer
				}
			}
		}
		if !existPair {
			return trades[i]
		}
	}
	return nil
}

func getPercentageToTrades(percents []int, amounts []*TokenAmount, pairs []Pair, currencyOut *Token, options *BestSmartTradeOptions) (map[int][]*Trade, error) {
	percentToTrades := make(map[int][]*Trade)
	for i, percent := range percents {
		amount := amounts[i]

		trades, err := BestTradeExactIn(pairs, amount, currencyOut, &options.BestTradeOptions, nil, amount, nil)
		if err != nil {
			return nil, err
		}
		percentToTrades[percent] = trades
	}
	return percentToTrades, nil
}

func getAmountDistribution(currencyAmountOut *TokenAmount, splitPercentage int) ([]int, []*TokenAmount, error) {
	percents := make([]int, 0)
	amounts := make([]*TokenAmount, 0)
	inputAmount := currencyAmountOut.Raw()
	for i := 1; i <= 100/splitPercentage; i++ {
		percents = append(percents, i*splitPercentage)
		currencyAmount, err := NewCurrencyAmount(currencyAmountOut.Currency, new(big.Int).Div(
			new(big.Int).Mul(inputAmount, big.NewInt(int64(i*splitPercentage))),
			constants.B100,
		))
		if err != nil {
			return nil, nil, err
		}
		amounts = append(amounts, &TokenAmount{
			CurrencyAmount: currencyAmount,
			Token:          currencyAmountOut.Token,
		})
	}
	return percents, amounts, nil
}

// extension of the input output comparator that also considers other dimensions of the trade in ranking them
func SmartTradeComparator(a, b *SmartTrade) int {
	return InputOutputComparator(a, b)
}
