package api

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/teambition/gear"

	"github.com/yiwen-ai/wallet-api/src/bll"
	"github.com/yiwen-ai/wallet-api/src/conf"
	"github.com/yiwen-ai/wallet-api/src/logging"
	"github.com/yiwen-ai/wallet-api/src/util"
)

// NewApp ...
func NewApp() *gear.App {
	app := gear.New()

	app.Set(gear.SetTrustedProxy, true)
	app.Set(gear.SetBodyParser, &bodyParser{gear.DefaultBodyParser(2 << 18)}) // 512kb
	// ignore TLS handshake error
	app.Set(gear.SetLogger, log.New(gear.DefaultFilterWriter(), "", 0))
	app.Set(gear.SetCompress, gear.ThresholdCompress(128))
	app.Set(gear.SetGraceTimeout, time.Duration(conf.Config.Server.GracefulShutdown)*time.Second)
	app.Set(gear.SetSender, &sendObject{})
	app.Set(gear.SetEnv, conf.Config.Env)

	app.UseHandler(logging.AccessLogger)
	err := util.DigInvoke(func(blls *bll.Blls, routers []*gear.Router) error {
		for _, router := range routers {
			app.UseHandler(router)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		headers := http.Header{}
		headers.Set("x-auth-user", util.JARVIS.String())
		headers.Set("x-auth-app", util.JARVIS.String())
		ctxHeader := util.ContextHTTPHeader(headers)
		ctx = gear.CtxWith[util.ContextHTTPHeader](ctx, &ctxHeader)
		if err := blls.Walletbase.InitApp(ctx, app); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logging.Panicf("DigInvoke error: %v", err)
	}

	return app
}

type bodyParser struct {
	inner gear.BodyParser
}

func (d *bodyParser) MaxBytes() int64 {
	return d.inner.MaxBytes()
}

func (d *bodyParser) Parse(buf []byte, body any, mediaType, charset string) error {
	if len(buf) == 0 {
		return gear.ErrBadRequest.WithMsg("request entity empty")
	}

	if strings.HasPrefix(mediaType, gear.MIMEApplicationCBOR) {
		return cbor.Unmarshal(buf, body)
	}

	return d.inner.Parse(buf, body, mediaType, charset)
}

type sendObject struct{}

func (s *sendObject) Send(ctx *gear.Context, code int, data any) error {
	if strings.HasPrefix(ctx.GetHeader(gear.HeaderAccept), gear.MIMEApplicationCBOR) || strings.HasPrefix(ctx.GetHeader(gear.HeaderContentType), gear.MIMEApplicationCBOR) {
		data, err := cbor.Marshal(data)
		if err != nil {
			return ctx.Error(err)
		}
		ctx.Type(gear.MIMEApplicationCBOR)
		return ctx.End(code, data)
	}

	return ctx.JSON(code, data)
}
