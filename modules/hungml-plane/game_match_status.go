package hungml_plane

type EGameMatchStatus int

const (
	MatchStatusWaitJoin EGameMatchStatus = iota
	MatchStatusReady
	MatchStatusRunning
	MatchStatusEnd
)
