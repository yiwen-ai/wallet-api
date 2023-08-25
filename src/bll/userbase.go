package bll

import (
	"context"

	"github.com/yiwen-ai/wallet-api/src/logging"
	"github.com/yiwen-ai/wallet-api/src/service"
	"github.com/yiwen-ai/wallet-api/src/util"
)

type Userbase struct {
	svc service.APIHost
}

type IDs struct {
	IDs []util.ID `json:"ids" cbor:"ids"`
}

func (b *Userbase) LoadUserInfo(ctx context.Context, ids ...util.ID) []UserInfo {
	output := SuccessResponse[[]UserInfo]{}
	if len(ids) == 0 {
		return []UserInfo{}
	}

	ids = util.RemoveDuplicates(ids)
	err := b.svc.Post(ctx, "/v1/user/batch_get_info", IDs{ids}, &output)
	if err != nil {
		logging.Warningf("Userbase.LoadUserInfo error: %v", err)
		return []UserInfo{}
	}

	return output.Result
}
