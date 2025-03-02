package process

import (
	"context"
	"cw/config"
	"cw/models"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"golang.org/x/sync/errgroup"
)

func WithdrawFactory(addresses []string) ([]models.WithdrawAction, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("Нет списка адресов.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	var (
		result = make([]models.WithdrawAction, len(addresses))
		mu     sync.Mutex
	)

	for i, address := range addresses {
		address := address
		g.Go(func() error {
			action := withdrawActionInit(address)

			mu.Lock()
			result[i] = *action
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func withdrawActionInit(address string) *models.WithdrawAction {
	chain := getRandomChain(config.WithdrawCfg.Chain)
	currency := getRandomChain(config.WithdrawCfg.Currency)
	amount := getRandomAmount(config.WithdrawCfg.AmountRange)
	time := getRandomAmount(config.WithdrawCfg.TimeRange)

	return &models.WithdrawAction{
		Address:   address,
		CEX:       config.WithdrawCfg.CEX,
		Chain:     chain,
		Currency:  currency,
		Amount:    amount,
		TimeRange: time,
	}
}

func getRandomChain(chains []string) string {
	return chains[rand.Intn(len(chains))]
}

func getRandomAmount(amountArr []float64) float64 {
	switch len(amountArr) {
	case 0:
		return 0
	case 1:
		return amountArr[0]
	default:
		min, max := amountArr[0], amountArr[1]
		if min > max {
			min, max = max, min
		}

		if min == max {
			return min
		}

		randoValue := min + rand.Float64()*(max-min)
		return math.Round(randoValue*100) / 100
	}
}
