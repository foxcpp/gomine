package gomine

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

func (l *Lib) ShouldDownload() bool {
	return EvaluateRules(l.Rules)
}

func (l *Lib) Native() *Artifact {
	if runtime.GOOS == "linux" {
		return l.Downloads.NativesLinux
	}
	if runtime.GOOS == "darwin" {
		return l.Downloads.NativesMacOS
	}
	if runtime.GOOS == "windows" {
		return l.Downloads.NativesWin
	}
	return nil
}

// SavePath returns FS path where library should be stored when downloaded (path is relative to libraries directory root).
func (l *Lib) SavePath() (string, error) {
	// libraries/<package>/<name>/<version>/<name>-<version>.jar
	pkg, name, version, err := l.SplitName()
	if pkg == "" {
		return "", err
	}
	pkgPath := strings.Replace(pkg, ".", string(os.PathSeparator), -1)
	return filepath.Join(pkgPath, name, version, name + "-" + version + ".jar"), nil
}

// NativeSavePath returns FS path where library's "native" component should be stored when downloaded (path is
// relative to libraries directory root).
//
// If library have no native component - empty string is returned.
func (l *Lib) NativeSavePath() (string, error) {
	native := l.Native()
	if native == nil {
		return "", nil
	}

	// libraries/<package>/<name>/<version>/<name>-<version>-<native_string>.jar
	// If <native_string> contains ${arch} is should be replaced with "32" or "64".

	var nativeStr string
	if runtime.GOOS == "linux" {
		nativeStr = l.NativeSuffixes.Linux
	}
	if runtime.GOOS == "darwin" {
		nativeStr = l.NativeSuffixes.MacOS
	}
	if runtime.GOOS == "windows" {
		nativeStr = l.NativeSuffixes.Windows
	}
	if runtime.GOARCH == "386" { // TODO: check for other 32-bit archs
		nativeStr = strings.Replace(nativeStr, "${arch}", "32", -1)
	} else {
		nativeStr = strings.Replace(nativeStr, "${arch}", "64", -1)
	}

	pkg, name, version, err := l.SplitName()
	if err != nil {
		return "", err
	}
	pkgPath := strings.Replace(pkg, ".", string(os.PathSeparator), -1)
	return filepath.Join(pkgPath, name, version, name + "-" + version + "-" + nativeStr + ".jar"), nil
}

func (l *Lib) SplitName() (pkg, name, version string, err error) {
	splitten := strings.Split(l.Name, ":")
	if len(splitten) != 3 {
		return "", "", "", errors.New("malformed library name: " + l.Name)
	}
	// <package>:<name>:<version>
	return splitten[0], splitten[1], splitten[2], nil
}

func (v *Version) DownloadLibraries(libDir string) error {
	// TODO: download progress callback
	for _, lib := range v.Libraries {
		if !lib.ShouldDownload() {
			continue
		}

		// Download path is same as save path.
		// https://libraries.minecraft.net/<package>/<name>/<version>/<name>-<version>.jar
		path, err := lib.SavePath()
		if err != nil {
			return errors.Wrapf(err, "failed to get save path for %s", lib.Name)
		}
		if path != "" && lib.Downloads.MainJar != nil {
			if err := downloadLib(libDir, lib.Downloads.MainJar.URL, path, lib.Downloads.MainJar.SHA1); err != nil {
				return errors.Wrapf(err, "failed to download %s", lib.Name)
			}
		}

		nativePath, err := lib.NativeSavePath()
		if err != nil {
			return errors.Wrapf(err, "failed to get native save path for %s", lib.Name)
		}
		if nativePath != "" {
			if err := downloadLib(libDir, lib.Native().URL, nativePath, lib.Native().SHA1); err != nil {
				return errors.Wrapf(err, "failed to download natives for %s", lib.Name)
			}
		}
	}
	return nil
}

func downloadLib(libDir, libUrl, path, expectedHash string) error {
	fsPath := filepath.Join(libDir, path)
	if _, err := os.Stat(fsPath); !os.IsNotExist(err) {
		// File exists, skipping download.
		return nil
	}

	log.Println(libUrl)
	resp, err := http.Get(libUrl)
	if err != nil {
		return errors.Wrap(err, "failed to start download")
	}
	if resp.StatusCode != 200 {
		return errors.New("failed to start download: HTTP " + resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(fsPath), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create lib directory")
	}

	hasher := sha1.New()

	outFile, err := os.Create(fsPath + ".new")
	if err != nil {
		return errors.Wrap(err, "failed to open lib file for writting")
	}
	if _, err := io.Copy(io.MultiWriter(outFile, hasher), resp.Body); err != nil {
		os.Remove(fsPath + ".new")
		return errors.Wrap(err, "failed to download library")
	}

	if hex.EncodeToString(hasher.Sum([]byte{})) != expectedHash {
		os.Remove(fsPath + ".new")
		return errors.New("hash mismatch")
	}

	// TODO: Handle extraction rules.

	outFile.Close()
	resp.Body.Close()

	if err := os.Rename(fsPath + ".new", fsPath); err != nil {
		return errors.Wrap(err, "failed to rename lib file")
	}

	return nil
}