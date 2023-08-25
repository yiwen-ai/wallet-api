package api

import (
	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/middleware"
	"github.com/yiwen-ai/wallet-api/src/util"
)

func init() {
	util.DigProvide(newAPIs)
	util.DigProvide(newRouters)
}

// APIs ..
type APIs struct {
	Healthz     *Healthz
	Transaction *Transaction
	Wallet      *Wallet
}

func newAPIs(blls *bll.Blls) *APIs {
	return &APIs{
		Healthz:     &Healthz{blls},
		Transaction: &Transaction{blls},
		Wallet:      &Wallet{blls},
	}
}

func todo(ctx *gear.Context) error {
	return gear.ErrNotImplemented.WithMsgf("TODO: %s %s", ctx.Method, ctx.Path)
}

func newRouters(apis *APIs) []*gear.Router {

	router := gear.NewRouter()
	router.Get("/healthz", apis.Healthz.Get)

	// 允许匿名访问
	router.Get("/currencies", middleware.AuthAllowAnon.Auth, apis.Wallet.ListCurrencies)

	// access_token 访问
	router.Get("/v1/wallet", middleware.AuthToken.Auth, apis.Wallet.Get)
	router.Post("/v1/wallet/sponsor", middleware.AuthToken.Auth, apis.Wallet.Sponsor)

	router.Post("/v1/transaction/list_outgo", middleware.AuthToken.Auth, apis.Transaction.ListOutgo)
	router.Post("/v1/transaction/list_income", middleware.AuthToken.Auth, apis.Transaction.ListIncome)
	router.Post("/v1/transaction/list_shares", middleware.AuthToken.Auth, apis.Transaction.ListShares)

	return []*gear.Router{router}
}
