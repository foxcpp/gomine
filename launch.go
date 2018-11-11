package gomine

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (v *Version) BuildCommandLine(prof Profile, authData AuthData, versionsDir, libsDir, nativesDir, assetsDir string) (bin string, args []string, err error) {
	gameDir, err := filepath.Abs(prof.GameDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get abs path")
	}
	versionsDir, err = filepath.Abs(versionsDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get abs path")
	}
	libsDir, err = filepath.Abs(libsDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get abs path")
	}
	nativesDir, err = filepath.Abs(nativesDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get abs path")
	}
	assetsDir, err = filepath.Abs(assetsDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get abs path")
	}

	classPath, err := v.BuildClassPath(versionsDir, libsDir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to build classpath")
	}

	argsReplacer := strings.NewReplacer(
		"${natives_directory}", nativesDir,
		"${launcher_name}", "gomine-framework",
		"${launcher_version}", "0.1",
		"${classpath}", classPath,
		"${auth_player_name}", authData.PlayerName,
		"${version_name}", v.ID,
		"${game_directory}", gameDir,
		"${assets_root}", assetsDir,
		"${assets_index_name}", v.AssetIndex.ID,
		"${auth_uuid}", authData.UUID,
		"${auth_access_token}", authData.Token,
		"${user_type}", authData.UserType,
		"${version_type}", v.Type,
		"${resolution_width}", strconv.Itoa(prof.ResolutionWidth),
		"${resolution_height}", strconv.Itoa(prof.ResolutionHeight),
	)

	cmdLine := make([]string, 0, len(v.JVMArgs)+len(v.GameArgs)+10)

	javaBin := prof.JVMPath
	if javaBin == "" {
		javaBin, err = findSystemJava()
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to detect system java")
		}
	}

	for _, arg := range v.JVMArgs {
		if !EvaluateRules(arg.Rules, &prof) {
			continue
		}

		cmdLine = append(cmdLine, argsReplacer.Replace(arg.Value))
	}
	cmdLine = append(cmdLine, strings.Split(argsReplacer.Replace(prof.CustomJVMArgs), " ")...)
	if prof.HeapMaxMB != 0 {
		cmdLine = append(cmdLine, "-Xmx"+strconv.Itoa(prof.HeapMaxMB)+"M")
	}

	cmdLine = append(cmdLine, v.MainClass)

	for _, arg := range v.GameArgs {
		if !EvaluateRules(arg.Rules, &prof) {
			continue
		}

		cmdLine = append(cmdLine, argsReplacer.Replace(arg.Value))
	}
	cmdLine = append(cmdLine, strings.Split(argsReplacer.Replace(prof.CustomGameArgs), " ")...)

	return javaBin, cmdLine, nil
}

func (v *Version) BuildClassPath(versionDir, libsDir string) (string, error) {
	var pathSep string
	if runtime.GOOS == "windows" {
		pathSep = ";"
	} else {
		pathSep = ":"
	}

	libs := make([]string, 0, len(v.Libraries)+1)
	for _, lib := range v.Libraries {
		if !lib.ShouldUse() {
			continue
		}

		path, err := lib.SavePath()
		if err != nil {
			return "", errors.Wrapf(err, "failed to get library path for %s", lib.Name)
		}

		libs = append(libs, filepath.Join(libsDir, path))
	}
	libs = append(libs, filepath.Join(versionDir, v.ID, v.ID+".jar"))

	return strings.Join(libs, pathSep), nil
}
