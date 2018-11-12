//+build !linux,!windows,!darwin

package gomine

import "errors"

func OsVersion() (string, error) {
	return "unknown", errors.New("OsVersion: not implemented")
}