package hungml_plane

//code response
const (
	OpCodeResponseGameState int64 = iota
	OpCodeResponseError
)

//code request
const (
	OpCodeRequestGameActionReady int64 = iota
	OpCodeRequestBuyGun
	OpCodeRequestUpgradeGun
)
