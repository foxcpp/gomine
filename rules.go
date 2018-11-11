package gomine

import (
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
		if err != nil {
			return true
		}
		if !osMatched {
			return false
		}
	}
	// TODO: Check for OS version.
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
