package gomine

type Artifact struct {
	SHA1 string `json:"sha1"`
	Size uint64 `json:"size"`
	URL  string `json:"url"`
}

type RuleAct string

const (
	ActAllow    RuleAct = "allow"
	ActDisallow RuleAct = "disallow"
)

type Rule struct {
	Action RuleAct `json:"action"`
	OS     struct {
		Name          string `json:"name"`
		VersionRegexp string `json:"version"`
	} `json:"os"`
}

type AssetIndex struct {
	ID        string `json:"id"`
	TotalSize uint64 `json:"totalSize"`
	Artifact
}

type LibClassifiers struct {
	JavaDoc      *Artifact `json:"javadoc"`
	NativesLinux *Artifact `json:"natives-linux"`
	NativesMacOS *Artifact `json:"natives-osx"`
	NativesWin   *Artifact `json:"natives-windows"`
	Sources      *Artifact `json:"sources"`
}

type Lib struct {
	Downloads struct {
		MainJar *Artifact `json:"artifact"`
		LibClassifiers `json:"classifiers"`
	} `json:"downloads"`
	NativeSuffixes struct {
		Linux   string `json:"linux"`
		MacOS   string `json:"osx"`
		Windows string `json:"windows"`
	} `json:"natives"`
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

// TODO
type LogCfg struct {
	JVMArg string
	Type   string
	File   Artifact `json:"file"`
}

type Argument struct {
	Value string `json:"value"`
	Rules []Rule `json:"rules"`
}

type Version struct {
	AssetIndex AssetIndex `json:"assetIndex"`
	Downloads  struct {
		Client Artifact `json:"client"`
		Server Artifact `json:"server"`
	} `json:"downloads"`
	ID        string `json:"id"`
	Libraries []Lib  `json:"libraries"`
	//ClientLog LogCfg
	MainClass string `json:"mainClass"`
	GameArgs  []Argument
	JVMArgs   []Argument
	Type      string `json:"type"`
}