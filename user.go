package gomine

type Profile struct {
	VersionID string
	GameDir                           string
	JVMPath                           string
	HeapMaxMB                         int
	CustomJVMArgs                     string
	CustomGameArgs 					  string

	ResolutionWidth, ResolutionHeight int
}

type AuthData struct {
	UserType string
	PlayerName string
	UUID string
	Token string
}