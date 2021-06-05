package hungml_plane

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
)

func createGame(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var session *SessionContext
	var err error

	if session, err = unpackContext(ctx); err != nil {
		return "", err
	}

	request := struct {
		GameType int
	}{}

	if err = json.Unmarshal([]byte(payload), &request); err != nil {
		return "", err
	}

	if request.GameType != 1 {
		return "", errors.New("game type is not invalid")
	}

	matchID, err := nk.MatchCreate(ctx, fmt.Sprintf("%s-Game", "hungml"), map[string]interface{}{
		"creator": session.UserID,
		"type_id": request.GameType,
	})

	response := struct {
		MatchID string
	}{
		MatchID: matchID,
	}

	bs, err := json.Marshal(response)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

