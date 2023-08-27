package api

import (
	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/middleware"
)

type Wallet struct {
	blls *bll.Blls
}

func (a *Wallet) ListCurrencies(ctx *gear.Context) error {
	output, err := a.blls.Walletbase.ListCurrencies(ctx)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	return ctx.OkSend(bll.SuccessResponse[[]bll.Currency]{Result: output})
}

func (a *Wallet) Get(ctx *gear.Context) error {
	sess := gear.CtxValue[middleware.Session](ctx)

	output, err := a.blls.Walletbase.Get(ctx, sess.UserID)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	return ctx.OkSend(bll.SuccessResponse[*bll.WalletOutput]{Result: output})
}

func (a *Wallet) Sponsor(ctx *gear.Context) error {
	input := &bll.ExpendInput{}
	if err := ctx.ParseBody(input); err != nil {
		return err
	}

	sess := gear.CtxValue[middleware.Session](ctx)
	input.UID = &sess.UserID
	input.SubPayee = nil // TODO: not supported yet

	output, err := a.blls.Walletbase.Sponsor(ctx, input)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	if _, err = a.blls.Logbase.Log(ctx, bll.LogActionUserSponsor, 1, sess.UserID, &bll.Payload{
		Payer:    *input.UID,
		Payee:    &input.Payee,
		SubPayee: nil,
		Amount:   input.Amount,
	}); err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	return ctx.OkSend(bll.SuccessResponse[*bll.WalletOutput]{Result: output})
}
