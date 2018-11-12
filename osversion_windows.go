//+build windows

package gomine

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func OsVersion() (string, error) {
	ver, err := windows.GetVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d.%d", byte(ver), uint8(ver>>8), uint16(ver>>16)), nil
}

/*
TODO: That's how it should be implemented using non-deprecated winapi functions. But it doesn't works nor I
	  have desire to fix it right now.

var winlibHandle windows.Handle
var getFileVersionInfo, getFileVersionSize, verQueryValue uintptr
var versionStr string

func init() {
	var err error
	winlibHandle, err = windows.LoadLibrary("api-ms-win-core-version-l1-1-0.dll")
	if err != nil {
		panic(err)
	}

	getFileVersionInfo, err = windows.GetProcAddress(winlibHandle, "GetFileVersionInfo")
	if err != nil {
		panic(err)
	}

	getFileVersionSize, err = windows.GetProcAddress(winlibHandle, "GetFileVersionInfoSizeW")
	if err != nil {
		panic(err)
	}

	verQueryValue, err = windows.GetProcAddress(winlibHandle, "VerQueryValueW")
	if err != nil {
		panic(err)
	}
}

func OsVersion() (string, error) {
	if versionStr != "" {
		return versionStr, nil
	}

	filename, err := windows.UTF16PtrFromString("kernel32.dll")
	if err != nil {
		return "", err
	}

	fileinfoSize, _, err := syscall.Syscall(getFileVersionSize, 1, uintptr(unsafe.Pointer(filename)), 0, 0)
	if err != nil {
		return "", err
	}
	if fileinfoSize == 0 {
		return "", errors.New("osVersion: zero fileinfoSize")
	}

	fileinfoBuf := make([]byte, fileinfoSize)
	pBlock := uintptr(unsafe.Pointer(&fileinfoBuf[0]))

	res, _, err := syscall.Syscall6(getFileVersionInfo, 4, uintptr(unsafe.Pointer(filename)), 0, fileinfoSize, pBlock, 0, 0)
	if err != nil {
		return "", err
	}
	if res == 0 {
		return "", errors.New("osVersion: zero result of GetFileVersionInfo")
	}

	subblock, err := windows.UTF16PtrFromString("\\StringFileInfo\\\\ProductVersion")
	if err != nil {
		panic(err)
	}

	var resultPtr, resultLen uintptr

	res, _, err = syscall.Syscall6(verQueryValue, 4, pBlock, uintptr(unsafe.Pointer(subblock)), uintptr(unsafe.Pointer(&resultPtr)), uintptr(unsafe.Pointer(&resultLen)), 0, 0)
	if err != nil {
		return "", err
	}
	if res == 0 {
		return "", errors.New("osVersion: zero result of VerQueryValue")
	}

	utf16Str := (*(*[]uint16)(unsafe.Pointer(resultPtr)))[:resultLen]

	return windows.UTF16ToString(utf16Str), nil
}
*/