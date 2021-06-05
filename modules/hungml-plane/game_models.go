package hungml_plane

import "github.com/heroiclabs/nakama-common/runtime"

const (
	_DEFAULT_MAP_X = 10
	_DEFAULT_MAP_Y = 20

	_MAX_GUNNER = 20
)

type GameMatchState struct {
	Turn      int
	Player    PlayerStats
	Mobs      []Mob
	Bullets   []Bullet
	MapTarget map[string]*Mob

	Creator string
	Status  EGameMatchStatus
	Config  GameMatchConfig

	UserID string

	GameStepAction StepAction `json:"actions"`
	Presences      map[string]runtime.Presence
}

type MobConfig struct {
}

type GameMatchConfig struct {
	CurrentMobLevel int

	GoldToUpLevel    int
	TurnToUpLevelMob int

	DamageUpEachLevel int
	HpUpEachLevel     int
	DefUpEachLevel    int
}

type StepAction struct {
	Turn    int                `json:"turn"`
	Mobs    map[string]*Mob    `json:"mobs"`
	Bullets map[string]*Bullet `json:"bullets"`
	Gunners map[string]Gunner  `json:"gunners"`
	Player  PlayerStats        `json:"player"`

	ErrorText string `json:"error"`
}

type PlayerStats struct {
	Health int
	Def    int
	Gold   int

	Gunners []Gunner
}

type Gunner struct {
	ID       string `json:"id"`
	Type     int    `json:"type"`
	Damage   int
	LastTurn int
	Delay    int
	Level    int
	Velocity int

	GoldUpgradeRequired int `json:"gold_require"`
	QuantityRequired    int `json:"quantity_required"`

	//state
	IsDealDamage bool `json:"is_deal_damage"`
	IsRemoved    bool `json:"is_removed"`
	IsCreated    bool `json:"is_created"`
	IsUpgrade    bool `json:"is_upgrade"`
}

type Mob struct {
	ID       string `json:"id"`
	Damage   int
	Velocity int
	Hp       int `json:"hp"`
	Gold     int
	Def      int

	PosX int `json:"pox_x"`
	PosY int `json:"pos_y"`

	//state
	IsDealDamage bool `json:"is_deal_damage"`
	IsRemoved    bool `json:"is_removed"`
	IsCreated    bool `json:"is_created"`
}

type Bullet struct {
	ID          string `json:"id"`
	TargetMobID string
	Damage      int
	Velocity    int
	PosY        int

	//state
	IsDealDamage bool `json:"is_deal_damage"`
	IsRemoved    bool `json:"is_removed"`
	IsCreated    bool `json:"is_created"`
}

type SessionContext struct {
	UserID    string
	SessionID string
}

type RequestBuyGun struct {
	TypeID int `json:"type_id"`
}

type RequestUpgradeGun struct {
	ID          string   `json:"id"`
	ResourceIDs []string `json:"resource_ids"`
}
