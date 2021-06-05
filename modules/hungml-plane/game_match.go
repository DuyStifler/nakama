package hungml_plane

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/rs/xid"
)

type GameMatch struct {
	gameConfig *GameMatchConfig
}

type MessageComing struct {
	Code int    `json:"code"`
	Data []byte `json:"data"`
}

func (g GameMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule,
	params map[string]interface{}) (interface{}, int, string) {
	creator, ok := params["creator"].(string)
	if !ok {
		logger.Error("[MatchInit] error: error get creator on param")
		return nil, 0, ""
	}

	typeID, ok := params["game_type"].(int)
	if !ok {
		logger.Error("[MatchInit] error: error get game type")
		return nil, 0, ""
	}

	label := &GameMatchLabel{
		TypeID:  typeID,
		Members: nil,
	}

	labelBs, _ := json.Marshal(label)

	state := GameMatchState{
		Turn:      0,
		Creator:   creator,
		Status:    MatchStatusWaitJoin,
		Mobs:      []Mob{},
		MapTarget: make(map[string]*Mob),
		Presences: make(map[string]runtime.Presence),
		Player: PlayerStats{
			Health:  200,
			Def:     10,
			Gold:    0,
			Gunners: []Gunner{},
		}}

	return state, 1, string(labelBs)
}

func (g GameMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	gameState := state.(*GameMatchState)

	if gameState.Status != MatchStatusWaitJoin {
		return gameState, false, "have people joined"
	}

	return state, true, ""
}

func (g GameMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*GameMatchState)

	s.Status = MatchStatusReady
	s.Turn = 0

	for _, presence := range presences {
		if _, ok := s.Presences[presence.GetUserId()]; !ok {
			s.Presences[presence.GetUserId()] = presence
		}
	}

	return s
}

func (g GameMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*GameMatchState)

	s.Status = MatchStatusEnd

	return s
}

func (g GameMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	s := state.(*GameMatchState)
	s.GameStepAction = StepAction{
		Mobs:    make(map[string]*Mob),
		Bullets: make(map[string]*Bullet),
		Gunners: make(map[string]Gunner),
		Player: PlayerStats{
			Health: 0,
			Def:    0,
			Gold:   0,
		},
	}

	s.Turn++

	if s.Status == MatchStatusEnd {
		//TODO send message if len presences > 0

		//TODO stop game
		var needKick []runtime.Presence
		for _, presence := range s.Presences {
			needKick = append(needKick, presence)
		}
		_ = dispatcher.MatchKick(needKick)

		return nil
	}

	if s.Status == MatchStatusRunning {
		g.spawnMod(s)
		g.upgradeMob(s)
		g.moveMob(s)
		g.gunLockTarget(s)
	}

	for _, message := range messages {
		switch message.GetOpCode() {
		case OpCodeRequestGameActionReady:
			s.Status = MatchStatusRunning
		case OpCodeRequestBuyGun:
			g.HandleMessageComing(ctx, logger, db, nk, dispatcher, tick, state, messages, s)
		}
	}

	s.GameStepAction.Player.Health = s.Player.Health
	s.GameStepAction.Player.Gold = s.Player.Gold
	s.GameStepAction.Player.Def = s.Player.Def

	bs, _ := json.Marshal(s.GameStepAction)
	presences := []runtime.Presence{}

	for _, presence := range s.Presences {
		presences = append(presences, presence)
	}

	s.GameStepAction.Turn = s.Turn

	code := OpCodeResponseGameState
	if s.GameStepAction.ErrorText != "" {
		code = OpCodeResponseError
	}
	_ = dispatcher.BroadcastMessage(code, bs, presences, nil, false)

	return s
}

func (g GameMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	panic("implement me")
}

func (g GameMatch) HandleMessageComing(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData, gameState *GameMatchState) {
	for _, message := range messages {
		var messageComing MessageComing
		_ = json.Unmarshal(message.GetData(), &messageComing)

		switch message.GetOpCode() {
		case OpCodeRequestBuyGun:
			if len(gameState.Player.Gunners) >= _MAX_GUNNER {
				return
			}

			req := RequestBuyGun{}
			err := json.Unmarshal([]byte(messageComing.Data), &req)
			if err != nil {
				logger.Error("[BUY GUN] ", err.Error())
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Buy Gun] json unmarsal %s ", err.Error())
				return
			}

			//TODO assign type id

			//TODO fake
			if gameState.Player.Gold < 100 {
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Buy Gun]  gold id not enought")
				return
			}

			gunner := Gunner{
				ID:                  xid.New().String(),
				Type:                1,
				Damage:              20,
				LastTurn:            0,
				Delay:               2,
				Level:               1,
				Velocity:            2,
				GoldUpgradeRequired: 2000,
			}
			gameState.Player.Gunners = append(gameState.Player.Gunners, gunner)
			gunner.IsCreated = true

			if gameState.Player.Gold < 10 {
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Buy Gun] not enough gold")
				return
			}

			gameState.Player.Gold -= 100
			gameState.GameStepAction.Gunners[gunner.ID] = gunner
		case OpCodeRequestUpgradeGun:
			req := RequestUpgradeGun{}
			err := json.Unmarshal(messageComing.Data, &req)
			if err != nil {
				logger.Error("[UPGRADE GUN] ", err.Error())
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Upgrade Gun] json unmarsal %s ", err.Error())
				return
			}

			rIdx := []int{}
			gunnerIdx := 0
			for idx := range gameState.Player.Gunners {
				if gameState.Player.Gunners[idx].ID == req.ID {
					gunnerIdx = idx
				}

				for _, resID := range req.ResourceIDs {
					if resID == gameState.Player.Gunners[idx].ID {
						rIdx = append(rIdx, idx)
						goto ContinueLoop
					}
				}

			ContinueLoop:
			}

			gunner := gameState.Player.Gunners[gunnerIdx]
			if gunner.QuantityRequired > len(rIdx) {
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Upgrade Gun] quantity is not enough")
				return
			}

			if gameState.Player.Gold < gunner.GoldUpgradeRequired {
				gameState.GameStepAction.ErrorText = fmt.Sprintf("[Upgrade Gun] gold is not enough")
				return
			}

			for idx := range rIdx {
				if gameState.Player.Gunners[idx].Type != gunner.Type {
					gameState.GameStepAction.ErrorText = fmt.Sprintf("[Upgrade Gun] type is not the same")
					return
				}
			}

			for idx := range rIdx {
				gameState.Player.Gunners = append(gameState.Player.Gunners[:idx], gameState.Player.Gunners[idx+1:]...)
			}

			//TODO fake
			gunner.Type++
			gunner.Damage += 20
			gunner.Velocity++
			gunner.QuantityRequired++
			gunner.GoldUpgradeRequired += 20000


			gameState.GameStepAction.Gunners[gunner.ID] = gunner
		}
	}
}

func (g GameMatch) spawnMod(s *GameMatchState) {
	mob := Mob{
		ID:   xid.New().String(),
		PosY: 0,
		PosX: rand.Intn(10),
	}

	mob.Def = s.Config.CurrentMobLevel*s.Config.DamageUpEachLevel + 20
	mob.Hp = s.Config.CurrentMobLevel*s.Config.HpUpEachLevel + 20
	mob.Damage = s.Config.CurrentMobLevel*s.Config.DamageUpEachLevel + 20

	s.Mobs = append(s.Mobs, mob)

	//insert mob to step data
	s.GameStepAction.Mobs[mob.ID] = &mob
}

func (g GameMatch) upgradeMob(s *GameMatchState) {
	if s.Turn*s.Config.TurnToUpLevelMob == 0 {
		s.Config.CurrentMobLevel++
	}
}

func (g GameMatch) moveMob(s *GameMatchState) {
	for idx := range s.Mobs {
		s.Mobs[idx].PosY += s.Mobs[idx].Velocity
		if s.Mobs[idx].PosY == _DEFAULT_MAP_Y {
			//TODO damage player
			s.Player.Health -= s.Mobs[idx].Damage
			s.GameStepAction.Player.Health = s.Player.Health

			//add to response
			if _, ok := s.GameStepAction.Mobs[s.Mobs[idx].ID]; !ok {
				s.GameStepAction.Mobs[s.Mobs[idx].ID] = &Mob{
					IsDealDamage: true,
					IsRemoved:    true,
				}
			}

			//delete mob
			s.Mobs = append(s.Mobs[:idx], s.Mobs[idx+1:]...)

			if s.Player.Health <= 0 {
				s.Status = MatchStatusEnd
				//TODO stop game
			}
		}
	}
}

func (g GameMatch) gunLockTarget(s *GameMatchState) {
	for _, gun := range s.Player.Gunners {
		idx := g.doGetRandomMob(s)
		mob := s.Mobs[idx]

		bullet := Bullet{
			ID:          xid.New().String(),
			TargetMobID: mob.ID,
			Damage:      gun.Damage,
			Velocity:    gun.Velocity,
			PosY:        _DEFAULT_MAP_Y,
		}

		if _, ok := s.MapTarget[bullet.TargetMobID]; !ok {
			s.MapTarget[bullet.TargetMobID] = &mob
		}
	}
}

func (g GameMatch) doBulletAction(s *GameMatchState) {
	for idx := range s.Bullets {
		s.Bullets[idx].PosY -= s.Bullets[idx].Velocity
		mobTarget := s.MapTarget[s.Bullets[idx].TargetMobID]
		if mobTarget.PosY >= s.Bullets[idx].PosY {
			//TODO damage mob
			damage := s.Bullets[idx].Damage - mobTarget.Def
			if damage <= 0 {
				damage = 1
			}
			mobTarget.Hp -= damage

			//add to response
			if _, ok := s.GameStepAction.Bullets[s.Bullets[idx].ID]; !ok {
				s.GameStepAction.Bullets[s.Bullets[idx].ID] = &Bullet{
					IsDealDamage: true,
					IsRemoved:    true,
				}
			}

			if mobTarget.Hp <= 0 {
				//TODO mob is destroyed
				s.GameStepAction.Mobs[mobTarget.ID].IsRemoved = true
			}
		}
	}
}

func (g GameMatch) doGetRandomMob(s *GameMatchState) int {
	return rand.Intn(len(s.Mobs))
}
