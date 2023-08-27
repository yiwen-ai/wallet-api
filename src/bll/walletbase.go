package bll

import (
	"context"
	"math"

	"github.com/teambition/gear"
	"github.com/yiwen-ai/wallet-api/src/service"
	"github.com/yiwen-ai/wallet-api/src/util"
)

type Walletbase struct {
	svc service.APIHost
}

type Currency struct {
	Name      string `json:"name" cbor:"name"`
	Alpha     string `json:"alpha" cbor:"alpha"`
	Decimals  uint8  `json:"decimals" cbor:"decimals"`
	Code      uint16 `json:"code" cbor:"code"`
	MinAmount uint   `json:"min_amount" cbor:"min_amount"`
	MaxAmount uint   `json:"max_amount" cbor:"max_amount"`
}

func (b *Walletbase) ListCurrencies(ctx context.Context) ([]Currency, error) {
	output := SuccessResponse[[]Currency]{}
	if err := b.svc.Get(ctx, "/currencies", &output); err != nil {
		return nil, err
	}

	return output.Result, nil
}

type WalletOutput struct {
	Sequence uint64   `json:"sequence" cbor:"sequence"`
	Award    int64    `json:"award" cbor:"award"`
	Topup    int64    `json:"topup" cbor:"topup"`
	Income   int64    `json:"income" cbor:"income"`
	Credits  uint64   `json:"credits" cbor:"credits"`
	Level    uint8    `json:"level" cbor:"level"`
	Txn      *util.ID `json:"txn,omitempty" cbor:"txn,omitempty"`
}

func (w *WalletOutput) SetLevel() {
	if w.Credits > 0 {
		w.Level = uint8(math.Floor(math.Log10(float64(w.Credits))))
	}
}

func (b *Walletbase) Get(ctx context.Context, uid util.ID) (*WalletOutput, error) {
	output := SuccessResponse[WalletOutput]{}
	if err := b.svc.Get(ctx, "/v1/wallet?uid="+uid.String(), &output); err != nil {
		return nil, err
	}
	output.Result.SetLevel()
	return &output.Result, nil
}

type ExpendInput struct {
	Payee       util.ID     `json:"payee" cbor:"payee"`
	Amount      int64       `json:"amount" cbor:"amount" validate:"gte=1,lte=1000000"`
	UID         *util.ID    `json:"uid" cbor:"uid"`                                 // 非客户端参数
	SubPayee    *util.ID    `json:"sub_payee,omitempty" cbor:"sub_payee,omitempty"` // 非客户端参数
	Description string      `json:"description,omitempty" cbor:"description,omitempty"`
	Payload     *util.Bytes `json:"payload,omitempty" cbor:"payload,omitempty"`
}

func (i *ExpendInput) Validate() error {
	if err := util.Validator.Struct(i); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	return nil
}

func (b *Walletbase) Sponsor(ctx context.Context, input *ExpendInput) (*WalletOutput, error) {
	output := SuccessResponse[WalletOutput]{}
	if err := b.svc.Post(ctx, "/v1/wallet/sponsor", input, &output); err != nil {
		return nil, err
	}

	output.Result.SetLevel()
	return &output.Result, nil
}

type TransactionOutput struct {
	ID           util.ID     `json:"id" cbor:"id"`
	Sequence     int64       `json:"sequence" cbor:"sequence"`
	Payee        *util.ID    `json:"payee,omitempty" cbor:"payee,omitempty"`
	SubPayee     *util.ID    `json:"sub_payee,omitempty" cbor:"sub_payee,omitempty"`
	Status       int8        `json:"status" cbor:"status"`
	Kind         string      `json:"kind" cbor:"kind"`
	Amount       int64       `json:"amount" cbor:"amount"`
	SysFee       int64       `json:"sys_fee" cbor:"sys_fee"`
	SubShares    int64       `json:"sub_shares" cbor:"sub_shares"`
	Description  string      `json:"description,omitempty" cbor:"description,omitempty"`
	Payload      *util.Bytes `json:"payload,omitempty" cbor:"payload,omitempty"`
	PayeeInfo    *UserInfo   `json:"payee_info,omitempty" cbor:"payee_info,omitempty"`
	SubPayeeInfo *UserInfo   `json:"sub_payee_info,omitempty" cbor:"sub_payee_info,omitempty"`
}

type Transactions []TransactionOutput

func (list *Transactions) LoadUsers(loader func(ids ...util.ID) []UserInfo) {
	if len(*list) == 0 {
		return
	}

	ids := make([]util.ID, 0, len(*list))
	for _, v := range *list {
		if v.Payee != nil {
			ids = append(ids, *v.Payee)
		}
	}
	for _, v := range *list {
		if v.SubPayee != nil {
			ids = append(ids, *v.SubPayee)
		}
	}

	users := loader(ids...)
	if len(users) == 0 {
		return
	}

	infoMap := make(map[util.ID]*UserInfo, len(users))
	for i := range users {
		infoMap[*users[i].ID] = &users[i]
		infoMap[*users[i].ID].ID = nil
	}

	for i := range *list {
		v := (*list)[i]
		if v.Payee != nil {
			v.PayeeInfo = infoMap[*v.Payee]
			v.Payee = nil
		}
		if v.SubPayee != nil {
			v.SubPayeeInfo = infoMap[*v.SubPayee]
			v.SubPayee = nil
		}
	}
}

func (b *Walletbase) ListOutgo(ctx context.Context, input *UIDPagination) (*SuccessResponse[Transactions], error) {
	output := SuccessResponse[Transactions]{}
	if err := b.svc.Post(ctx, "/v1/transaction/list_outgo", input, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (b *Walletbase) ListIncome(ctx context.Context, input *UIDPagination) (*SuccessResponse[Transactions], error) {
	output := SuccessResponse[Transactions]{}
	if err := b.svc.Post(ctx, "/v1/transaction/list_income", input, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (b *Walletbase) ListShares(ctx context.Context, input *UIDPagination) (*SuccessResponse[Transactions], error) {
	output := SuccessResponse[Transactions]{}
	if err := b.svc.Post(ctx, "/v1/transaction/list_shares", input, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
