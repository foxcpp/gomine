package gomine

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const versionsUrl = `https://launchermeta.mojang.com/mc/game/version_manifest.json`

type Root struct {
	LauncherDir string
	AuthData    AuthData

	LatestRelease  string
	LatestSnapshot string
	knownVersions  map[string]VersionMeta
	versions       map[string]Version
}

type VersionMeta struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`

	Installed bool
}

func UseRoot(path string) Root {
	return Root{LauncherDir: path}
}

func (r *Root) VersionsDir() string {
	return filepath.Join(r.LauncherDir, "versions")
}

func (r *Root) AssetsDir() string {
	return filepath.Join(r.LauncherDir, "assets")
}

func (r *Root) LibrariesDir() string {
	return filepath.Join(r.LauncherDir, "libraries")
}

func (r *Root) GetVersion(id string) (*Version, error) {
	if r.knownVersions == nil {
		if _, err := r.Versions(); err != nil {
			return nil, err
		}
	}
	if r.versions == nil {
		r.versions = make(map[string]Version)
	}

	meta, prs := r.knownVersions[id]
	if !prs {
		return nil, errors.New("update: unknown version id")
	}

	var versionInfo *Version
	if !meta.Installed {
		if meta.URL == "" {
			return nil, errors.New("update: can't download local-only version")
		}

		resp, err := http.Get(meta.URL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to download version info")
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("update: HTTP " + resp.Status)
		}
		blob, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to download version info")
		}
		resp.Body.Close()

		versionDir := filepath.Join(r.VersionsDir(), meta.ID)
		if err := os.MkdirAll(versionDir, os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "failed to create version directory")
		}

		if err := ioutil.WriteFile(filepath.Join(versionDir, meta.ID+".json"), blob, os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "failed to write version info")
		}

		if versionInfo, err = ReadVersionJSON(blob); err != nil {
			return nil, errors.Wrap(err, "failed to parse version info")
		}
	} else {
		versionDir := filepath.Join(r.VersionsDir(), meta.ID)
		blob, err := ioutil.ReadFile(filepath.Join(versionDir, meta.ID+".json"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to write version info")
		}

		if versionInfo, err = ReadVersionJSON(blob); err != nil {
			return nil, errors.Wrap(err, "failed to parse version info")
		}
	}

	r.versions[meta.ID] = *versionInfo
	return versionInfo, nil
}

func (r *Root) UpdateVersion(ver *Version) error {
	if err := ver.DownloadLibraries(r.LibrariesDir()); err != nil {
		return err
	}
	if err := ver.DownloadAssetsIndex(r.AssetsDir()); err != nil {
		return err
	}
	if err := ver.DownloadAssets(r.AssetsDir()); err != nil {
		return err
	}
	if err := ver.DownloadClient(r.VersionsDir()); err != nil {
		return err
	}
	return nil
}

func (r *Root) RunVersion(ver *Version, prof *Profile, logRedirect io.Writer) error {
	nativesDir, err := ioutil.TempDir("", "gomine")
	if err != nil {
		return err
	}
	defer os.RemoveAll(nativesDir)

	if err := ver.ExtractNatives(r.LibrariesDir(), nativesDir); err != nil {
		return err
	}

	bin, args, err := ver.BuildCommandLine(*prof, r.AuthData, r.VersionsDir(), r.LibrariesDir(), nativesDir, r.AssetsDir())
	if err != nil {
		return err
	}
	//log.Println("Command line:", bin, args)
	cmd := exec.Command(bin, args...)
	if logRedirect == nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, logRedirect)
		cmd.Stderr = io.MultiWriter(os.Stderr, logRedirect)
	}
	return cmd.Run()
}

func (r *Root) Versions() (map[string]VersionMeta, error) {
	versions := make(map[string]VersionMeta)
	remoteVers, err := r.RemoteVersions()
	if err != nil {
		return nil, err
	}
	for _, ver := range remoteVers.Versions {
		versions[ver.ID] = ver
	}
	r.LatestRelease = remoteVers.Latest.Release
	r.LatestSnapshot = remoteVers.Latest.Snapshot

	versionDir, err := ioutil.ReadDir(r.VersionsDir())
	for _, localVer := range versionDir {
		if _, prs := versions[localVer.Name()]; prs {
			s := versions[localVer.Name()]
			s.Installed = true
			versions[localVer.Name()] = s
		} else {
			versions[localVer.Name()] = VersionMeta{
				ID:        localVer.Name(),
				Type:      "local",
				Installed: true,
			}
		}
	}

	r.knownVersions = versions
	return versions, nil
}

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []VersionMeta `json:"versions"`
}

func (r *Root) RemoteVersions() (*VersionManifest, error) {
	resp, err := http.Get(versionsUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download versions manifest")
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("failed to download versions manifest: HTTP " + resp.Status)
	}
	var out VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, errors.Wrap(err, "failed to decode versions manifest")
	}
	return &out, nil
}
