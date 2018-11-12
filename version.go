package gomine

type Asset struct {
	Hash string `json:"hash"`
	Size uint64 `json:"size"`
}

type AssetIndexContents struct {
	Objects map[string]Asset `json:"objects"`
}

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
	Action RuleAct `json:"action" mapstructure:"action"`
	// OS information. All fields are regexes.
	OS     struct {
		Name    string `json:"name" mapstructure:"name"`
		Version string `json:"version" mapstructure:"version"`
		Arch    string `json:"arch" mapstructure:"arch"`
	} `json:"os" mapstructure:"os"`
	Features struct {
		IsDemoUser *bool `json:"is_demo_user" mapstructure:"is_demo_user"`
		HasCustomResolution *bool `json:"has_custom_resolution" mapstructure:"has_custom_resolution"`
	} `json:"features" mapstructure:"features"`
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
		MainJar        *Artifact `json:"artifact"`
		LibClassifiers `json:"classifiers"`
	} `json:"downloads"`
	NativeSuffixes struct {
		Linux   string `json:"linux"`
		MacOS   string `json:"osx"`
		Windows string `json:"windows"`
	} `json:"natives"`
	Name         string `json:"name"`
	Rules        []Rule `json:"rules"`
	ExtractRules struct {
		Exclude []string `json:"exclude"`
	} `json:"extract"`
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
