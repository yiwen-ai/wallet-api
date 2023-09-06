package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/price"
	"github.com/stripe/stripe-go/v75/webhook"
	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/conf"
	"github.com/yiwen-ai/wallet-api/src/logging"
	"github.com/yiwen-ai/wallet-api/src/middleware"
	"github.com/yiwen-ai/wallet-api/src/util"
)

func init() {
	stripe.Key = conf.Config.Stripe.SecretKey
}

type Checkout struct {
	blls *bll.Blls
	cfg  conf.Stripe
}

type CheckoutConfig struct {
	Provider   string `json:"provider" cbor:"provider"`
	PublicKey  string `json:"public_key" cbor:"public_key"`
	UnitAmount int64  `json:"unit_amount" cbor:"unit_amount"`
	Currency   string `json:"currency" cbor:"currency"`
}

func (a *Checkout) GetConfig(ctx *gear.Context) error {
	p, err := price.Get(
		a.cfg.PriceID,
		nil,
	)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	return ctx.OkJSON(bll.SuccessResponse[CheckoutConfig]{Result: CheckoutConfig{
		Provider:   "stripe",
		PublicKey:  a.cfg.PubKey,
		UnitAmount: p.UnitAmount,
		Currency:   string(p.Currency),
	}})
}

func (a *Checkout) Get(ctx *gear.Context) error {
	input := &bll.QueryId{}
	if err := ctx.ParseURL(input); err != nil {
		return err
	}
	sess := gear.CtxValue[middleware.Session](ctx)

	output, err := a.blls.Walletbase.GetCharge(ctx, sess.UserID, input.ID, input.Fields)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}
	output.ChargePayload = nil
	return ctx.OkSend(bll.SuccessResponse[*bll.ChargeOutput]{Result: output})
}

type CheckoutInput struct {
	Quantity uint    `json:"quantity" cbor:"quantity" validate:"gte=50,lte=1000000"`
	Currency *string `json:"currency" cbor:"currency"`
}

func (i *CheckoutInput) Validate() error {
	if err := util.Validator.Struct(i); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	return nil
}

type CheckoutOutput struct {
	ID         util.ID `json:"id" cbor:"id"`
	PaymentURL string  `json:"payment_url" cbor:"payment_url"`
}

func (a *Checkout) Create(ctx *gear.Context) error {
	input := &CheckoutInput{}
	if err := ctx.ParseBody(input); err != nil {
		return err
	}

	if input.Currency != nil {
		input.Currency = util.Ptr(strings.ToLower(*input.Currency))
		if err := a.blls.Walletbase.Currencies.Validate(*input.Currency); err != nil {
			return err
		}
	}

	sess := gear.CtxValue[middleware.Session](ctx)
	output, err := a.blls.Walletbase.CreateCharge(ctx, &bll.ChargeInput{
		UID:      sess.UserID,
		Provider: "stripe",
		Quantity: input.Quantity,
	})
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	logging.SetTo(ctx, "chargeId", output.ID.String())
	err = a.createSession(ctx, output.ID, &stripe.CheckoutSessionParams{
		SuccessURL:       util.Ptr(a.cfg.SuccessUrl),
		Mode:             util.Ptr(string(stripe.CheckoutSessionModePayment)),
		Currency:         input.Currency,
		CustomerCreation: util.Ptr("always"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: util.Ptr(int64(input.Quantity)),
				Price:    util.Ptr(a.cfg.PriceID),
			},
		},
		Metadata: map[string]string{
			"uid": sess.UserID.String(),
			"cid": output.ID.String(),
		},
	})

	if err != nil {
		logging.SetTo(ctx, "createSessionError", err.Error())
	}
	return err
}

func (a *Checkout) createSession(ctx *gear.Context, chargeID util.ID, params *stripe.CheckoutSessionParams) error {
	sess := gear.CtxValue[middleware.Session](ctx)
	if customer, _ := a.blls.Walletbase.GetCustomer(ctx, sess.UserID, "stripe", util.Ptr("customer")); customer != nil {
		params.Customer = stripe.String(customer.Customer)
		params.CustomerCreation = nil
	}
	cs, err := session.New(params)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	logging.SetTo(ctx, "checkoutId", cs.ID)
	payload, err := cbor.Marshal(cs)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	_, err = a.blls.Walletbase.UpdateCharge(ctx, &bll.UpdateChargeInput{
		UID:           sess.UserID,
		ID:            chargeID,
		CurrentStatus: 0,
		Status:        1,
		Currency:      util.Ptr(string(cs.Currency)),
		Amount:        util.Ptr(uint(cs.AmountTotal)),
		ChargeID:      util.Ptr(cs.ID),
		ChargePayload: util.Ptr(util.Bytes(payload)),
	})
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	return ctx.OkSend(bll.SuccessResponse[CheckoutOutput]{Result: CheckoutOutput{
		ID:         chargeID,
		PaymentURL: cs.URL,
	}})
}

func (a *Checkout) ListCharges(ctx *gear.Context) error {
	input := &bll.UIDPagination{}
	if err := ctx.ParseBody(input); err != nil {
		return err
	}
	sess := gear.CtxValue[middleware.Session](ctx)
	input.UID = &sess.UserID

	output, err := a.blls.Walletbase.ListCharges(ctx, input)
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	// for i := range output {
	// 	output[i].UID = nil
	// }

	return ctx.OkSend(bll.SuccessResponse[[]bll.ChargeOutput]{Result: output})
}

func (a *Checkout) StripeWebhook(ctx *gear.Context) error {
	b, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		return gear.ErrBadRequest.WithMsgf("read body failed: %v", err)
	}

	event, err := webhook.ConstructEvent(b, ctx.Req.Header.Get("Stripe-Signature"), a.cfg.WebhookKey)
	if err != nil {
		return gear.ErrBadRequest.WithMsgf("webhook.ConstructEvent failed: %v", err)
	}

	logging.SetTo(ctx, "eventType", event.Type)
	logging.SetTo(ctx, "eventId", event.ID)

	var obj map[string]interface{}
	if event.Data != nil && len(event.Data.Object) > 0 {
		obj = event.Data.Object
	}

	logging.SetTo(ctx, "objectType", obj["object"])
	logging.SetTo(ctx, "objectId", obj["id"])

	if event.Type == "checkout.session.completed" {
		if err = a.completeSession(ctx, event.Data.Raw); err != nil {
			logging.SetTo(ctx, "completeSessionError", err.Error())
			return ctx.Error(err)
		}
	}

	return ctx.OkJSON(bll.SuccessResponse[bool]{Result: true})
}

func (a *Checkout) completeSession(ctx *gear.Context, data []byte) error {
	cs := &stripe.CheckoutSession{}
	if err := json.Unmarshal(data, cs); err != nil {
		return gear.ErrBadRequest.WithMsgf("json.Unmarshal failed: %v", err)
	}
	uid, err := util.ParseID(cs.Metadata["uid"])
	if err != nil {
		return gear.ErrBadRequest.WithMsgf("parse uid failed: %v", err)
	}

	logging.SetTo(ctx, "uid", uid)
	cid, err := util.ParseID(cs.Metadata["cid"])
	if err != nil {
		return gear.ErrBadRequest.WithMsgf("parse uid failed: %v", err)
	}

	logging.SetTo(ctx, "chargeId", cid)
	headers := http.Header{}
	headers.Set("x-auth-user", uid.String())
	headers.Set("x-auth-app", util.JARVIS.String())
	headers.Set("x-real-ip", ctx.GetHeader("x-real-ip"))
	headers.Set("x-request-id", ctx.GetHeader("x-request-id"))

	ctxHeader := util.ContextHTTPHeader(headers)
	cctx := gear.CtxWith[util.ContextHTTPHeader](ctx, &ctxHeader)

	charge, err := a.blls.Walletbase.CompleteCharge(cctx, &bll.CompleteChargeInput{
		UID:           uid,
		ID:            cid,
		Currency:      string(cs.Currency),
		Amount:        uint(cs.AmountTotal),
		ChargeID:      cs.ID,
		ChargePayload: util.Bytes(data),
	})

	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}

	if cs.Customer != nil && cs.CustomerDetails != nil {
		data, err := cbor.Marshal(cs.CustomerDetails)
		if err == nil {
			_, err = a.blls.Walletbase.UpsertCustomer(cctx, &bll.CustomerInput{
				UID:      uid,
				Provider: "stripe",
				Customer: cs.Customer.ID,
				Payload:  util.Bytes(data),
			})
		}
		if err != nil {
			logging.SetTo(ctx, "upsertCustomerError", err.Error())
		} else {
			logging.SetTo(ctx, "customer", cs.Customer.ID)
		}
	}

	if _, err = a.blls.Logbase.Log(ctx, bll.LogActionUserTopup, 1, uid, &bll.Payload{
		Kind:   "charge",
		ID:     charge.ID,
		Amount: int64(charge.Quantity),
	}); err != nil {
		logging.SetTo(ctx, "writeLogError", err.Error())
	}

	return nil
}
