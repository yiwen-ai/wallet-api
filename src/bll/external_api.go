package bll

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/teambition/gear"
	"github.com/yiwen-ai/wallet-api/src/logging"
	"github.com/yiwen-ai/wallet-api/src/service"
	"github.com/yiwen-ai/wallet-api/src/util"
)

var userAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"

type ExternalAPI struct {
	redis *service.Redis
	rates atomic.Pointer[ExchangeRatesOutput]
}

type ExchangeRatesOutput struct {
	LastUpdate uint               `json:"last_update" cbor:"last_update"` // unix timestamp in seconds
	NextUpdate uint               `json:"next_update" cbor:"next_update"`
	Base       string             `json:"base" cbor:"base"` // base should be "HKD"
	Rates      map[string]float32 `json:"rates" cbor:"rates"`
}

func (b *ExternalAPI) ExchangeRate(ctx context.Context) (*ExchangeRatesOutput, error) {
	v := b.rates.Load()
	if v == nil {
		v := &ExchangeRatesOutput{}
		_ = b.redis.GetCBOR(ctx, "exchange_rates", &v)
	}

	r := util.Int63n(3500) + 100
	if v != nil && time.Now().Unix()-r < int64(v.LastUpdate) {
		return v, nil
	}

	if v != nil && v.LastUpdate > 0 {
		// we should update later
		go b.exchangeRate(ctx)
		return v, nil
	}

	// we should update now:
	return b.exchangeRate(ctx)
}

func (b *ExternalAPI) exchangeRate(ctx context.Context) (rate *ExchangeRatesOutput, err error) {
	defer func() {
		if err != nil {
			logging.Logger.Err(logging.Log{
				"action": "fetch_exchange_rates",
				"error":  err.Error(),
			})
		}
	}()

	// https://www.exchangerate-api.com/docs
	h := http.Header{}
	h.Set("User-Agent", userAgent)
	ctx = gear.CtxWith[util.CtxHeader](ctx, util.Ptr(util.CtxHeader(h)))

	type exchangeRateOutput struct {
		Result     string `json:"result"`
		LastUpdate uint   `json:"time_last_update_unix"`
		NextUpdate uint   `json:"time_next_update_unix"`
		// should be "HKD"
		Base  string             `json:"base_code"`
		Rates map[string]float32 `json:"conversion_rates"`
	}

	output := &exchangeRateOutput{}
	api := "https://v6.exchangerate-api.com/v6/245ef0a5e7b4a1799b2d9a64/latest/HKD"
	err = util.RequestJSON(ctx, util.ExternalHTTPClient, http.MethodGet, api, nil, output)
	if err != nil {
		return nil, err
	}

	if output.Result != "success" {
		err = errors.New("fetch exchange rate failed")
		return nil, err
	}

	rate = &ExchangeRatesOutput{
		LastUpdate: output.LastUpdate,
		NextUpdate: output.NextUpdate,
		Base:       output.Base,
		Rates:      output.Rates,
	}

	b.rates.Store(rate)
	if err = b.redis.SetCBOR(ctx, "exchange_rates", rate, 0); err != nil {
		return nil, err
	}

	logging.Logger.Info(logging.Log{
		"action":     "fetch_exchange_rates",
		"base":       rate.Base,
		"lastUpdate": time.Unix(int64(rate.LastUpdate), 0),
		"nextUpdate": time.Unix(int64(rate.NextUpdate), 0),
	})

	return rate, nil
}
