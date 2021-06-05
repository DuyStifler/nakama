package hungml_plane

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := Register(initializer); err != nil {
		return err
	}
	return nil
}


func Register(initializer runtime.Initializer) error {
	createMatchFunc := func(ctx context.Context, logger runtime.Logger, db *sql.DB, nr runtime.NakamaModule) (runtime.Match, error) {
		return &GameMatch{
			gameConfig: &GameMatchConfig{
				CurrentMobLevel:   0,
				GoldToUpLevel:     20,
				TurnToUpLevelMob:  5,
				DamageUpEachLevel: 5,
				HpUpEachLevel:     5,
				DefUpEachLevel:    5,
			},
		}, nil
	}

	if err := initializer.RegisterMatch("hungml-game", createMatchFunc); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(fmt.Sprintf("%s-%s", "hungml", "CreateMatchPlane"), createGame); err != nil {
		return err
	}

	return nil
}
