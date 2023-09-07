package api

import (
	"net/http"

	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/conf"
	"github.com/yiwen-ai/wallet-api/src/middleware"
	"github.com/yiwen-ai/wallet-api/src/util"
)

func init() {
	util.DigProvide(newAPIs)
	util.DigProvide(newRouters)
}

// APIs ..
type APIs struct {
	Checkout    *Checkout
	Healthz     *Healthz
	Transaction *Transaction
	Wallet      *Wallet
}

func newAPIs(blls *bll.Blls) *APIs {
	return &APIs{
		Checkout:    &Checkout{blls: blls, cfg: conf.Config.Stripe},
		Healthz:     &Healthz{blls},
		Transaction: &Transaction{blls},
		Wallet:      &Wallet{blls},
	}
}

func newRouters(apis *APIs) []*gear.Router {

	router := gear.NewRouter()
	router.Use(func(ctx *gear.Context) error {
		h := http.Header{}
		// inject headers into context for base service
		util.CopyHeader(h, ctx.Req.Header,
			"x-real-ip",
			"x-request-id",
		)

		ctx.WithContext(gear.CtxWith[util.CtxHeader](ctx.Context(), util.Ptr(util.CtxHeader(h))))
		return nil
	})

	router.Get("/healthz", apis.Healthz.Get)

	// 允许匿名访问
	router.Get("/currencies", middleware.AuthAllowAnon.Auth, apis.Wallet.ListCurrencies)

	// access_token 访问
	router.Get("/v1/wallet", middleware.AuthToken.Auth, apis.Wallet.Get)
	router.Post("/v1/wallet/sponsor", middleware.AuthToken.Auth, apis.Wallet.Sponsor)

	router.Post("/v1/transaction/list_outgo", middleware.AuthToken.Auth, apis.Transaction.ListOutgo)
	router.Post("/v1/transaction/list_income", middleware.AuthToken.Auth, apis.Transaction.ListIncome)
	router.Post("/v1/transaction/list_shares", middleware.AuthToken.Auth, apis.Transaction.ListShares)

	router.Get("/v1/checkout/config", middleware.AuthToken.Auth, apis.Checkout.GetConfig)
	router.Get("/v1/checkout", middleware.AuthToken.Auth, apis.Checkout.Get)
	router.Post("/v1/checkout", middleware.AuthToken.Auth, apis.Checkout.Create)
	router.Post("/v1/checkout/list", middleware.AuthToken.Auth, apis.Checkout.ListCharges)

	router.Post("/v1/webhook/stripe", apis.Checkout.StripeWebhook)

	return []*gear.Router{router}
}
