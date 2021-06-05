package hungml_plane

import (
	"context"
	"errors"
	"github.com/heroiclabs/nakama-common/runtime"
)

func unpackContext(ctx context.Context) (*SessionContext, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil, errors.New("user not found")
	}

	sessionID, ok := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if !ok {
		return nil, errors.New("session not found")
	}

	return &SessionContext{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}
