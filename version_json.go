package gomine

import (
	"encoding/json"
	"errors"
	"strings"
)

var ErrIncompatibleFormat = errors.New("ReadVersionJSON: JSON format version is higher than supported by library")
var ErrInvalidFormat = errors.New("ReadVersionJSON: passed JSON object doesn't matches expected schema")

var defaultJVMArgs = []Argument{
	{Value: "-Xss1M"},
	{Value: "-Djava.library.path=${natives_directory}"},
	{Value: "-Dminecraft.launcher.brand=${launcher_name}"},
	{Value: "-Dminecraft.launcher.version=${launcher_version}"},
	{Value: "-cp"},
	{Value: "${classpath}"},
	{Value: "-Xmx2G"},
	{Value: "-XX:+UnlockExperimentalVMOptions"},
	{Value: "-XX:+UseG1GC"},
	{Value: "-XX:G1NewSizePercent=20"},
	{Value: "-XX:G1ReservePercent=20"},
	{Value: "-XX:MaxGCPauseMillis=50"},
	{Value: "-XX:G1HeapRegionSize=32M"},
}

type versionJson struct {
	Version
	MinimumLauncherVersion uint   `json:"minimumLauncherVersion"`
	MinecraftArgs          string `json:"minecraftArguments"`
	Logging                struct {
		Client LogCfg `json:"client"`
	}
	Arguments struct {
		Game []interface{} `json:"game"`
		JVM  []interface{} `json:"jvm"`
	} `json:"arguments"`
}

func processRawArgs(raw []interface{}) ([]Argument, error) {
	res := []Argument{}
	for _, arg := range raw {
		switch arg.(type) {
		case string:
			res = append(res, Argument{Value: arg.(string)})
		case map[string]interface{}:
			mapArg := arg.(map[string]interface{})
			saneArg := Argument{}

			switch mapArg["value"].(type) {
			case string:
				saneArg.Value = mapArg["value"].(string)
			case []interface{}:
				strArr := []string{}
				for _, rawVal := range mapArg["value"].([]interface{}) {
					str, ok := rawVal.(string)
					if !ok {
						return res, ErrInvalidFormat
					}
					strArr = append(strArr, str)
				}

				saneArg.Value = strings.Join(strArr, " ")
			default:
				return res, ErrInvalidFormat
			}

			res = append(res, saneArg)
		default:
			return res, ErrInvalidFormat
		}
	}
	return res, nil
}

func ReadVersionJSON(in []byte) (*Version, error) {
	raw := versionJson{}
	var err error

	if err = json.Unmarshal(in, &raw); err != nil {
		return nil, err
	}

	if raw.MinimumLauncherVersion > 21 {
		return &raw.Version, err
	}

	if raw.Arguments.Game != nil {
		raw.Version.GameArgs, err = processRawArgs(raw.Arguments.Game)
		if err != nil {
			return nil, err
		}
	}
	if raw.Arguments.JVM != nil {
		raw.Version.JVMArgs, err = processRawArgs(raw.Arguments.JVM)
		if err != nil {
			return nil, err
		}
	}

	if raw.MinecraftArgs != "" {
		raw.Version.GameArgs = []Argument{}
		for _, arg := range strings.Split(raw.MinecraftArgs, " ") {
			raw.Version.GameArgs = append(raw.Version.GameArgs, Argument{Value: arg})
		}
	}

	if raw.Version.JVMArgs == nil || len(raw.Version.JVMArgs) == 0 {
		raw.Version.JVMArgs = defaultJVMArgs
	}

	return &raw.Version, nil
}
