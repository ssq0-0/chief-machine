package modules

import (
	"context"
	"cw/httpClient"
	"cw/models"
	"sync"

	"golang.org/x/sync/errgroup"
)

type ModulesFasad interface {
	Withdraw(token, address, network string, amount float64) error
	GetBalances(token string) error
	GetPrices(token string) error
}

type ModuleFactory func(cfg *models.CexConfig) (ModulesFasad, error)

func ModulesInit(cfg *models.CexConfig) (map[string]ModulesFasad, error) {
	hc, err := httpClient.NewHttpClient(
		httpClient.WithHttp2(),
		httpClient.WithProxy(""),
	)
	if err != nil {
		return nil, err
	}
	modules := map[string]ModuleFactory{
		"bybit": func(cfg *models.CexConfig) (ModulesFasad, error) {
			return NewBybitModule(
				cfg.BybitCfg.BalanceEndpoint,
				cfg.BybitCfg.TickersEndpoint,
				cfg.BybitCfg.API_key,
				cfg.BybitCfg.API_secret,
				hc,
			)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

	var (
		mu     sync.Mutex
		result = make(map[string]ModulesFasad, len(modules))
	)

	for name, factory := range modules {
		name, factory := name, factory

		g.Go(func() error {
			module, err := factory(cfg)
			if err != nil {
				return err
			}
			mu.Lock()
			result[name] = module
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}
