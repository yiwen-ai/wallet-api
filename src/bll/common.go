package bll

import (
	"context"
	"strings"

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
	ExternalAPI *ExternalAPI
	Logbase     *Logbase
	Taskbase    *Taskbase
	Userbase    *Userbase
	Walletbase  *Walletbase
}

// NewBlls ...
func NewBlls(redis *service.Redis) *Blls {
	cfg := conf.Config.Base
	return &Blls{
		ExternalAPI: &ExternalAPI{redis: redis},
		Logbase:     &Logbase{svc: service.APIHost(cfg.Logbase)},
		Taskbase:    &Taskbase{svc: service.APIHost(cfg.Taskbase)},
		Userbase:    &Userbase{svc: service.APIHost(cfg.Userbase)},
		Walletbase:  &Walletbase{svc: service.APIHost(cfg.Walletbase)},
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
	Status    *int8       `json:"status,omitempty" cbor:"status,omitempty"`
	Fields    *[]string   `json:"fields,omitempty" cbor:"fields,omitempty"`
}

func (i *UIDPagination) Validate() error {
	if err := util.Validator.Struct(i); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	return nil
}

type Payload struct {
	Kind     string   `json:"kind" cbor:"kind"`
	ID       util.ID  `json:"id" cbor:"id"`
	Payer    util.ID  `json:"payer" cbor:"payer"`
	Payee    *util.ID `json:"payee,omitempty" cbor:"payee,omitempty"`
	SubPayee *util.ID `json:"sub_payee,omitempty" cbor:"sub_payee,omitempty"`
	Amount   int64    `json:"amount" cbor:"amount"`
}

type QueryId struct {
	ID     util.ID `json:"id,omitempty" cbor:"id,omitempty" query:"id"`
	Fields *string `json:"fields,omitempty" cbor:"fields,omitempty" query:"fields"`
}

func (i *QueryId) Validate() error {
	return nil
}

type Currency struct {
	Name     string  `json:"name" cbor:"name"`
	Alpha    string  `json:"alpha" cbor:"alpha"`
	Decimals uint8   `json:"decimals" cbor:"decimals"`
	Code     uint16  `json:"code" cbor:"code"`
	Rate     float32 `json:"exchange_rate" cbor:"exchange_rate"` // HKD: 10000
}

type Currencies []Currency

func (cs Currencies) Validate(cur string) error {
	cur = strings.ToUpper(cur)
	for _, c := range cs {
		if c.Alpha == cur {
			return nil
		}
	}

	return gear.ErrBadRequest.From(gear.ErrBadRequest.WithMsgf("currency %s not supported", cur))
}
