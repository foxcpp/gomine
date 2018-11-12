//+build linux

package gomine

import "golang.org/x/sys/unix"

func OsVersion() (string, error) {
	buf := unix.Utsname{}
	if err := unix.Uname(&buf); err != nil {
		return "", err
	}

	return string(buf.Release[:]), nil
}