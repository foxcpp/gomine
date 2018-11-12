//+build darwin

package gomine

// XXX: UNTESTED!

/*
#include <errno.h>
#include <sys/sysctl.h>

int readOsrelease(char* outBuf, int bufSize) {
	res = sysctlbyname("kern.osrelease", outbuf, &bufsize, NULL, 0);
	if (res < 0) {
		return res;
	}
	return bufSize;
}
 */
import "C"

func OsVersion() (string, error) {
	buf := C.CBytes(make([]byte, 256))
	ret, err := C.readOsrelease(buf)
	if ret == 0 {
		return "", err
	}
	return C.GoStringN(buf, ret)
}