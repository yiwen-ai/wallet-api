package bll

import (
	"context"

	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/conf"
	"github.com/yiwen-ai/wallet-api/src/service"
	"github.com/yiwen-ai/wallet-api/src/util"
)

func init() {
	util.DigProvide(NewBlls)
}

// Blls ...
type Blls struct {
	Logbase    *Logbase
	Taskbase   *Taskbase
	Userbase   *Userbase
	Walletbase *Walletbase
}

// NewBlls ...
func NewBlls() *Blls {
	cfg := conf.Config.Base
	return &Blls{
		Logbase:    &Logbase{svc: service.APIHost(cfg.Logbase)},
		Taskbase:   &Taskbase{svc: service.APIHost(cfg.Taskbase)},
		Userbase:   &Userbase{svc: service.APIHost(cfg.Userbase)},
		Walletbase: &Walletbase{svc: service.APIHost(cfg.Walletbase)},
	}
}

func (b *Blls) Stats(ctx context.Context) (res map[string]any, err error) {
	return b.Userbase.svc.Stats(ctx)
}

type SuccessResponse[T any] struct {
	Retry         int        `json:"retry,omitempty" cbor:"retry,omitempty"`
	TotalSize     int        `json:"total_size,omitempty" cbor:"total_size,omitempty"`
	NextPageToken util.Bytes `json:"next_page_token,omitempty" cbor:"next_page_token,omitempty"`
	Job           string     `json:"job,omitempty" cbor:"job,omitempty"`
	Result        T          `json:"result" cbor:"result"`
}

type UserInfo struct {
	ID      *util.ID `json:"id,omitempty" cbor:"id,omitempty"` // should clear this field when return to client
	CN      string   `json:"cn" cbor:"cn"`
	Name    string   `json:"name" cbor:"name"`
	Picture string   `json:"picture" cbor:"picture"`
	Status  int8     `json:"status" cbor:"status"`
	Kind    int8     `json:"kind" cbor:"kind"`
}

type Pagination struct {
	PageToken *util.Bytes `json:"page_token,omitempty" cbor:"page_token,omitempty"`
	PageSize  *uint16     `json:"page_size,omitempty" cbor:"page_size,omitempty" validate:"omitempty,gte=5,lte=100"`
	Fields    *[]string   `json:"fields,omitempty" cbor:"fields,omitempty"`
}

func (i *Pagination) Validate() error {
	if err := util.Validator.Struct(i); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	return nil
}

type UIDPagination struct {
	UID       *util.ID    `json:"uid" cbor:"uid"` // 非客户端参数
	PageToken *util.Bytes `json:"page_token,omitempty" cbor:"page_token,omitempty"`
	PageSize  *uint16     `json:"page_size,omitempty" cbor:"page_size,omitempty" validate:"omitempty,gte=5,lte=100"`
	Kind      *string     `json:"kind,omitempty" cbor:"kind,omitempty"`
	Fields    *[]string   `json:"fields,omitempty" cbor:"fields,omitempty"`
}

func (i *UIDPagination) Validate() error {
	if err := util.Validator.Struct(i); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	return nil
}

type Payload struct {
	Payer    util.ID  `json:"payer" cbor:"payer"`
	Payee    util.ID  `json:"payee" cbor:"payee"`
	SubPayee *util.ID `json:"sub_payee,omitempty" cbor:"sub_payee,omitempty"`
	Amount   int64    `json:"amount" cbor:"amount"`
}

type QueryIdCn struct {
	ID     *util.ID `json:"id,omitempty" cbor:"id,omitempty" query:"id"`
	CN     *string  `json:"cn,omitempty" cbor:"cn,omitempty" query:"cn"`
	Fields *string  `json:"fields,omitempty" cbor:"fields,omitempty" query:"fields"`
}

func (i *QueryIdCn) Validate() error {
	if i.ID == nil && i.CN == nil {
		return gear.ErrBadRequest.WithMsg("id or cn is required")
	}
	return nil
}
