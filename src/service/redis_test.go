// (c) 2022-present, Yiwen AI, LLC. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yiwen-ai/wallet-api/src/util"
)

func TestRedis(t *testing.T) {
	cli := NewRedis()
	t.Run("GetCBOR and SetCBOR", func(t *testing.T) {
		assert := assert.New(t)
		type testUser struct {
			ID   util.ID `json:"id" cbor:"id"`
			Name string  `json:"name" cbor:"name"`
		}

		ctx := context.Background()

		var user testUser
		err := cli.GetCBOR(ctx, "test", &user)
		assert.True(util.IsNotFoundErr(err))

		user = testUser{
			ID:   util.NewID(),
			Name: "tester",
		}
		err = cli.SetCBOR(ctx, "test", &user, 1)
		assert.NoError(err)

		var user2 testUser
		err = cli.GetCBOR(ctx, "test", &user2)
		assert.NoError(err)
		assert.Equal(user.ID, user2.ID)
		assert.Equal("tester", user2.Name)

		time.Sleep(1100 * time.Millisecond)
		err = cli.GetCBOR(ctx, "test", &user2)
		assert.True(util.IsNotFoundErr(err))
	})
}
