package bll

import (
	"context"
	"math"
	"net/url"

	"github.com/teambition/gear"
	"github.com/yiwen-ai/wallet-api/src/service"
	"github.com/yiwen-ai/wallet-api/src/util"
)

type Walletbase struct {
	svc        service.APIHost
	Currencies Currencies
}

func (b *Walletbase) InitApp(ctx context.Context, _ *gear.App) error {
	output, err := b.listCurrencies(ctx)
	if err != nil {
		return err
	}
	b.Currencies = output
	return nil
}

func (b *Walletbase) listCurrencies(ctx context.Context) ([]Currency, error) {
	output := SuccessResponse[[]Currency]{}
	if err := b.svc.Get(ctx, "/currencies", &output); err != nil {
		return nil, err
	}

	return output.Result, nil
}

type WalletOutput struct {
	Sequence uint64  `json:"sequence" cbor:"sequence"`
	Award    int64   `json:"award" cbor:"award"`
	Topup    int64   `json:"topup" cbor:"topup"`
	Income   int64   `json:"income" cbor:"income"`
	Credits  uint64  `json:"credits" cbor:"credits"`
	Level    uint8   `json:"level" cbor:"level"`
	Txn      util.ID `json:"txn" cbor:"txn"`
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
	Payer        *util.ID    `json:"payer,omitempty" cbor:"payer,omitempty"`
	Payee        *util.ID    `json:"payee,omitempty" cbor:"payee,omitempty"`
	SubPayee     *util.ID    `json:"sub_payee,omitempty" cbor:"sub_payee,omitempty"`
	Status       int8        `json:"status" cbor:"status"`
	Kind         string      `json:"kind" cbor:"kind"`
	Amount       int64       `json:"amount" cbor:"amount"`
	SysFee       int64       `json:"sys_fee" cbor:"sys_fee"`
	SubShares    int64       `json:"sub_shares" cbor:"sub_shares"`
	CreatedAt    int64       `json:"created_at" cbor:"created_at"`
	Description  string      `json:"description,omitempty" cbor:"description,omitempty"`
	Payload      *util.Bytes `json:"payload,omitempty" cbor:"payload,omitempty"`
	PayerInfo    *UserInfo   `json:"payer_info,omitempty" cbor:"payer_info,omitempty"`
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
		if v.Payer != nil {
			ids = append(ids, *v.Payer)
		}
	}
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
		if v.Payer != nil {
			v.PayerInfo = infoMap[*v.Payer]
			v.Payer = nil
		}
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

	for i := range output.Result {
		output.Result[i].CreatedAt = output.Result[i].ID.UnixMs()
	}
	return &output, nil
}

func (b *Walletbase) ListIncome(ctx context.Context, input *UIDPagination) (*SuccessResponse[Transactions], error) {
	output := SuccessResponse[Transactions]{}
	if err := b.svc.Post(ctx, "/v1/transaction/list_income", input, &output); err != nil {
		return nil, err
	}

	for i := range output.Result {
		output.Result[i].CreatedAt = output.Result[i].ID.UnixMs()
	}
	return &output, nil
}

func (b *Walletbase) ListShares(ctx context.Context, input *UIDPagination) (*SuccessResponse[Transactions], error) {
	output := SuccessResponse[Transactions]{}
	if err := b.svc.Post(ctx, "/v1/transaction/list_shares", input, &output); err != nil {
		return nil, err
	}

	for i := range output.Result {
		output.Result[i].CreatedAt = output.Result[i].ID.UnixMs()
	}
	return &output, nil
}

type CustomerInput struct {
	UID      util.ID    `json:"uid" cbor:"uid"`
	Provider string     `json:"provider" cbor:"provider"`
	Customer string     `json:"customer" cbor:"customer"`
	Payload  util.Bytes `json:"payload" cbor:"payload"`
}

type CustomerOutput struct {
	UID       util.ID     `json:"uid" cbor:"uid"`
	Provider  string      `json:"provider" cbor:"provider"`
	Customer  string      `json:"customer" cbor:"customer"`
	CreatedAt *int64      `json:"created_at,omitempty" cbor:"created_at,omitempty"`
	UpdatedAt *int64      `json:"updated_at,omitempty" cbor:"updated_at,omitempty"`
	Payload   *util.Bytes `json:"payload,omitempty" cbor:"payload,omitempty"`
	Customers []string    `json:"customers,omitempty" cbor:"customers,omitempty"`
}

func (b *Walletbase) GetCustomer(ctx context.Context, uid util.ID, provider string, fields *string) (*CustomerOutput, error) {
	output := SuccessResponse[CustomerOutput]{}

	query := url.Values{}
	query.Add("uid", uid.String())
	query.Add("provider", provider)
	if fields != nil && len(*fields) > 0 {
		query.Add("fields", *fields)
	}
	if err := b.svc.Get(ctx, "/v1/customer?"+query.Encode(), &output); err != nil {
		return nil, err
	}

	return &output.Result, nil
}

func (b *Walletbase) UpsertCustomer(ctx context.Context, input *CustomerInput) (*CustomerOutput, error) {
	output := SuccessResponse[CustomerOutput]{}
	if err := b.svc.Post(ctx, "/v1/customer", input, &output); err != nil {
		return nil, err
	}

	return &output.Result, nil
}

type ChargeInput struct {
	UID           util.ID     `json:"uid" cbor:"uid"`
	Provider      string      `json:"provider" cbor:"provider"`
	Quantity      uint        `json:"quantity" cbor:"quantity"`
	Currency      *string     `json:"currency,omitempty" cbor:"currency,omitempty"`
	Amount        *uint       `json:"amount,omitempty" cbor:"amount,omitempty"`
	ChargeID      *string     `json:"charge_id,omitempty" cbor:"charge_id,omitempty"`
	ChargePayload *util.Bytes `json:"charge_payload,omitempty" cbor:"charge_payload,omitempty"`
}

type UpdateChargeInput struct {
	UID           util.ID     `json:"uid" cbor:"uid"`
	ID            util.ID     `json:"id" cbor:"id"`
	CurrentStatus int8        `json:"current_status" cbor:"current_status"`
	Status        int8        `json:"status" cbor:"status"`
	Currency      *string     `json:"currency,omitempty" cbor:"currency,omitempty"`
	Amount        *uint       `json:"amount,omitempty" cbor:"amount,omitempty"`
	ChargeID      *string     `json:"charge_id,omitempty" cbor:"charge_id,omitempty"`
	ChargePayload *util.Bytes `json:"charge_payload,omitempty" cbor:"charge_payload,omitempty"`
	FailureCode   *string     `json:"failure_code,omitempty" cbor:"failure_code,omitempty"`
	FailureMsg    *string     `json:"failure_msg,omitempty" cbor:"failure_msg,omitempty"`
}

type CompleteChargeInput struct {
	UID           util.ID    `json:"uid" cbor:"uid"`
	ID            util.ID    `json:"id" cbor:"id"`
	Currency      string     `json:"currency" cbor:"currency"`
	Amount        uint       `json:"amount" cbor:"amount"`
	ChargeID      string     `json:"charge_id" cbor:"charge_id"`
	ChargePayload util.Bytes `json:"charge_payload" cbor:"charge_payload"`
}

type ChargeOutput struct {
	// UID       util.ID    `json:"uid" cbor:"uid"` // should not return to client
	ID             util.ID     `json:"id" cbor:"id"`
	Provider       string      `json:"provider" cbor:"provider"`
	Status         int8        `json:"status" cbor:"status"`
	Quantity       uint        `json:"quantity" cbor:"quantity"`
	CreatedAt      int64       `json:"created_at" cbor:"created_at"`
	UpdatedAt      *int64      `json:"updated_at,omitempty" cbor:"updated_at,omitempty"`
	ExpireAt       *int64      `json:"expire_at,omitempty" cbor:"expire_at,omitempty"`
	Currency       *string     `json:"currency,omitempty" cbor:"currency,omitempty"`
	Amount         *uint       `json:"amount,omitempty" cbor:"amount,omitempty"`
	AmountRefunded *uint       `json:"amount_refunded,omitempty" cbor:"amount_refunded,omitempty"`
	ChargeID       *string     `json:"charge_id,omitempty" cbor:"charge_id,omitempty"`
	ChargePayload  *util.Bytes `json:"charge_payload,omitempty" cbor:"charge_payload,omitempty"`
	Txn            *util.ID    `json:"txn,omitempty" cbor:"txn,omitempty"`
	TxnRefunded    *util.ID    `json:"txn_refunded,omitempty" cbor:"txn_refunded,omitempty"`
	FailureCode    *string     `json:"failure_code,omitempty" cbor:"failure_code,omitempty"`
	FailureMsg     *string     `json:"failure_msg,omitempty" cbor:"failure_msg,omitempty"`
}

func (b *Walletbase) GetCharge(ctx context.Context, uid, id util.ID, fields *string) (*ChargeOutput, error) {
	output := SuccessResponse[ChargeOutput]{}

	query := url.Values{}
	query.Add("uid", uid.String())
	query.Add("id", id.String())
	if fields != nil && len(*fields) > 0 {
		query.Add("fields", *fields)
	}
	if err := b.svc.Get(ctx, "/v1/charge?"+query.Encode(), &output); err != nil {
		return nil, err
	}

	output.Result.CreatedAt = output.Result.ID.UnixMs()
	return &output.Result, nil
}

func (b *Walletbase) CreateCharge(ctx context.Context, input *ChargeInput) (*ChargeOutput, error) {
	output := SuccessResponse[ChargeOutput]{}
	if err := b.svc.Post(ctx, "/v1/charge", input, &output); err != nil {
		return nil, err
	}

	return &output.Result, nil
}

func (b *Walletbase) UpdateCharge(ctx context.Context, input *UpdateChargeInput) (*ChargeOutput, error) {
	output := SuccessResponse[ChargeOutput]{}
	if err := b.svc.Patch(ctx, "/v1/charge", input, &output); err != nil {
		return nil, err
	}

	return &output.Result, nil
}

func (b *Walletbase) CompleteCharge(ctx context.Context, input *CompleteChargeInput) (*ChargeOutput, error) {
	output := SuccessResponse[ChargeOutput]{}
	if err := b.svc.Post(ctx, "/v1/charge/complete", input, &output); err != nil {
		return nil, err
	}

	return &output.Result, nil
}

func (b *Walletbase) ListCharges(ctx context.Context, input *UIDPagination) ([]ChargeOutput, error) {
	output := SuccessResponse[[]ChargeOutput]{}
	if err := b.svc.Post(ctx, "/v1/charge/list", input, &output); err != nil {
		return nil, err
	}

	for i := range output.Result {
		output.Result[i].CreatedAt = output.Result[i].ID.UnixMs()
	}
	return output.Result, nil
}

type CreditOutput struct {
	Txn         util.ID `json:"txn" cbor:"txn"`
	Kind        string  `json:"kind" cbor:"kind"`
	Amount      int64   `json:"amount" cbor:"amount"`
	CreatedAt   int64   `json:"created_at" cbor:"created_at"`
	Description string  `json:"description,omitempty" cbor:"description,omitempty"`
}

func (b *Walletbase) ListCredits(ctx context.Context, input *UIDPagination) ([]CreditOutput, error) {
	output := SuccessResponse[[]CreditOutput]{}
	if err := b.svc.Post(ctx, "/v1/wallet/list_credits", input, &output); err != nil {
		return nil, err
	}

	for i := range output.Result {
		output.Result[i].CreatedAt = output.Result[i].Txn.UnixMs()
	}
	return output.Result, nil
}
