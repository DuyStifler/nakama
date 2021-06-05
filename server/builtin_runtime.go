package server

import (
	"context"
	"database/sql"
	"github.com/heroiclabs/nakama-common/runtime"
	hungmlplane "github.com/heroiclabs/nakama/v3/modules/hungml-plane"
)

func doInitModule(ctx context.Context, runtimeLogger runtime.Logger, db *sql.DB, nk *RuntimeGoNakamaModule, initializer *RuntimeGoInitializer) error {
	var err error
	err = hungmlplane.InitModule(ctx, runtimeLogger, db, nk, initializer)
	if err != nil {
		return err
	}

	return nil
}
