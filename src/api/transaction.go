package api

import (
	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/middleware"
	"github.com/yiwen-ai/wallet-api/src/util"
)

type Transaction struct {
	blls *bll.Blls
}

func (a *Transaction) ListOutgo(ctx *gear.Context) error {
	input := &bll.UIDPagination{}
	if err := ctx.ParseBody(input); err != nil {
		return err
	}

	sess := gear.CtxValue[middleware.Session](ctx)
	input.UID = &sess.UserID

	output, err := a.blls.Walletbase.ListOutgo(ctx, input)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	output.Result.LoadUsers(func(ids ...util.ID) []bll.UserInfo {
		return a.blls.Userbase.LoadUserInfo(ctx, ids...)
	})

	return ctx.OkSend(output)
}

func (a *Transaction) ListIncome(ctx *gear.Context) error {
	input := &bll.UIDPagination{}
	if err := ctx.ParseBody(input); err != nil {
		return err
	}

	sess := gear.CtxValue[middleware.Session](ctx)
	input.UID = &sess.UserID

	output, err := a.blls.Walletbase.ListIncome(ctx, input)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	output.Result.LoadUsers(func(ids ...util.ID) []bll.UserInfo {
		return a.blls.Userbase.LoadUserInfo(ctx, ids...)
	})

	return ctx.OkSend(output)
}
