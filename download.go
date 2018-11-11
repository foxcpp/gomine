package gomine

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

func (l *Lib) ShouldUse() bool {
	return EvaluateRules(l.Rules, nil)
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
	return filepath.Join(pkgPath, name, version, name+"-"+version+".jar"), nil
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
	return filepath.Join(pkgPath, name, version, name+"-"+version+"-"+nativeStr+".jar"), nil
}

func (l *Lib) SplitName() (pkg, name, version string, err error) {
	splitten := strings.Split(l.Name, ":")
	if len(splitten) != 3 {
		return "", "", "", errors.New("malformed library name: " + l.Name)
	}
	// <package>:<name>:<version>
	return splitten[0], splitten[1], splitten[2], nil
}

func (l *Lib) ExtractNative(libDir, nativeDir string) error {
	path, err := l.NativeSavePath()
	if err != nil {
		return err
	}
	if path == "" {
		panic("attempt to extract non-existent native")
	}
	nativePath := filepath.Join(libDir, path)

	r, err := zip.OpenReader(nativePath)
	if err != nil {
		if os.IsNotExist(err) {
			panic("attempt to extract non-existent native")
		}
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		skip := false
		for _, exclude := range l.ExtractRules.Exclude {
			if strings.HasPrefix(file.Name, exclude) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return errors.Wrapf(err, "failed to extract %s", file.Name)
		}

		if err := os.MkdirAll(filepath.Dir(filepath.Join(nativeDir, file.Name)), os.ModePerm); err != nil {
			return errors.Wrap(err, "failed to create target dir")
		}

		out, err := os.Create(filepath.Join(nativeDir, file.Name))
		if err != nil {
			return errors.Wrapf(err, "failed to open %s for writting", file.Name)
		}

		if _, err := io.Copy(out, rc); err != nil {
			return errors.Wrapf(err, "failed to write %s", file.Name)
		}

		rc.Close()
		out.Close()
	}
	return nil
}

func (v *Version) DownloadLibraries(libDir string) error {
	// TODO: download progress callback
	for _, lib := range v.Libraries {
		if !lib.ShouldUse() {
			continue
		}

		// Download path is same as save path.
		// https://libraries.minecraft.net/<package>/<name>/<version>/<name>-<version>.jar
		path, err := lib.SavePath()
		if err != nil {
			return errors.Wrapf(err, "failed to get save path for %s", lib.Name)
		}
		if path != "" && lib.Downloads.MainJar != nil {
			if err := lib.Downloads.MainJar.Download(filepath.Join(libDir, path)); err != nil {
				return errors.Wrapf(err, "failed to download %s", lib.Name)
			}
		}

		nativePath, err := lib.NativeSavePath()
		if err != nil {
			return errors.Wrapf(err, "failed to get native save path for %s", lib.Name)
		}
		if nativePath != "" {
			if err := lib.Native().Download(filepath.Join(libDir, nativePath)); err != nil {
				return errors.Wrapf(err, "failed to download natives for %s", lib.Name)
			}
		}
	}
	return nil
}

func (v *Version) ExtractNatives(libDir, nativeDir string) error {
	for _, lib := range v.Libraries {
		if !lib.ShouldUse() || lib.Native() == nil {
			continue
		}
		savePath, _ := lib.NativeSavePath()
		log.Println("Extracting native libraries from", savePath+"...")
		if err := lib.ExtractNative(libDir, nativeDir); err != nil {
			return errors.Wrapf(err, "failed to extract natives for %s", lib.Name)
		}
	}
	return nil
}

func (v *Version) DownloadClient(versionsDir string) error {
	jarPath := filepath.Join(versionsDir, v.ID, v.ID+".jar")

	if err := os.MkdirAll(filepath.Dir(jarPath), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create version directory")
	}

	return v.Downloads.Client.Download(jarPath)
}

func (v *Version) DownloadAssetsIndex(assetsDir string) error {
	assetsIndxPath := filepath.Join(assetsDir, "indexes", v.AssetIndex.ID+".json")

	if err := os.MkdirAll(filepath.Dir(assetsIndxPath), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create version directory")
	}

	return v.AssetIndex.Download(assetsIndxPath)
}

func (v *Version) DownloadAssets(assetsDir string) error {
	assetsIndxPath := filepath.Join(assetsDir, "indexes", v.AssetIndex.ID+".json")

	assetsIndxBlob, err := ioutil.ReadFile(assetsIndxPath)
	if err != nil {
		return errors.Wrap(err, "failed to read assets index")
	}

	indx := AssetIndexContents{}
	if err := json.Unmarshal(assetsIndxBlob, &indx); err != nil {
		return errors.Wrap(err, "failed to parse assets index")
	}

	objectsDir := filepath.Join(assetsDir, "objects")
	if err := os.MkdirAll(objectsDir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create objects dir")
	}

	for path, asset := range indx.Objects {
		if err := asset.Download(objectsDir); err != nil {
			return errors.Wrapf(err, "failed to download asset %s", path)
		}
	}
	return nil
}

func (a *Asset) Download(objectsDir string) error {
	targetPath := filepath.Join(objectsDir, a.Hash[:2], a.Hash)
	url := "http://resources.download.minecraft.net/" + a.Hash[:2] + "/" + a.Hash
	return downloadAndCheck(targetPath, url, a.Hash)
	// TODO: Copy to virtual/legacy for pre-1.7.2 versions.
}

func (a *Artifact) Download(targetPath string) error {
	return downloadAndCheck(targetPath, a.URL, a.SHA1)
}

func downloadAndCheck(targetPath, url, expectedHash string) error {
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		hash := sha1.New()
		f, err := os.Open(targetPath)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}

		if _, err := io.Copy(hash, f); err != nil {
			return errors.Wrap(err, "failed to read file")
		}

		if hex.EncodeToString(hash.Sum([]byte{})) == expectedHash {
			return nil
		}
		// if existing file doesn't matches hash - redownload.
	}

	log.Println("Downloading", url+"...")
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to start download")
	}
	if resp.StatusCode != 200 {
		return errors.New("failed to start download: HTTP " + resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	hash := sha1.New()

	outFile, err := os.Create(targetPath + ".new")
	if err != nil {
		return errors.Wrap(err, "failed to open file for writting")
	}
	if _, err := io.Copy(io.MultiWriter(outFile, hash), resp.Body); err != nil {
		os.Remove(targetPath + ".new")
		return errors.Wrap(err, "failed to download file")
	}

	if hex.EncodeToString(hash.Sum([]byte{})) != expectedHash {
		os.Remove(targetPath + ".new")
		return errors.New("hash mismatch")
	}

	outFile.Close()
	resp.Body.Close()

	if err := os.Rename(targetPath+".new", targetPath); err != nil {
		return errors.Wrap(err, "failed to rename artifact file")
	}
	return nil
}
