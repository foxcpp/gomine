package gomine

import (
	"log"
	"regexp"
	"runtime"
)

func (r Rule) Applies(prof *Profile) bool {
	if r.OS.Name != "" {
		osName := runtime.GOOS
		if osName == "darwin" {
			osName = "osx"
		}

		osMatched, err := regexp.MatchString(r.OS.Name, osName)
		if err != nil || !osMatched {
			return false
		}
	}
	if r.OS.Version != "" {
		ver, err := OsVersion()
		if err != nil {
			log.Println("Failed to get OS version:", err)
			return false
		}

		verMatched, err := regexp.MatchString(r.OS.Version, ver)
		if err != nil || !verMatched {
			return false
		}
	}
	if r.OS.Arch != "" {
		// Should be checked against values of Java's os.arch values.
		// GOARCH=386   => os.arch=x86
		// GOARCH=amd64 => os.arch=amd64
		// TODO: other values?
		if runtime.GOARCH == "386" {
			return r.OS.Arch == "x86"
		}
		if runtime.GOARCH == "amd64" {
			return r.OS.Arch == "amd64"
		}
		return false
	}
	if r.Features.IsDemoUser != nil {
		return !(*r.Features.IsDemoUser)
	}
	if r.Features.HasCustomResolution != nil {
		return prof.ResolutionHeight != 0 && prof.ResolutionWidth != 0
	}
	return true
}

func EvaluateRules(rules []Rule, profile *Profile) bool {
	if rules == nil {
		return true
	}

	// TODO: Does it matches how official launcher evaluates them?
	last := ActDisallow
	for _, rule := range rules {
		if rule.Applies(profile) {
			last = rule.Action
		}
	}
	return last == ActAllow
}
