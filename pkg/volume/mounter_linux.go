// +build linux

/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"syscall"
)

type DiskMounter struct{}

func (mounter *DiskMounter) Mount(source string, target string, fstype string, flags string, data string) error {
	var sysflags uintptr
	if flags == "bind" {
		sysflags = syscall.MS_BIND
	}
	return syscall.Mount(source, target, fstype, sysflags, data)
}

func (mounter *DiskMounter) Unmount(target string, flags int) error {
	return syscall.Unmount(target, flags)
}

func (mounter *DiskMounter) RefCount(PD *PersistentDisk) (int, error) {
	contents, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return -1, err
	}
	refCount := 0
	deviceName := path.Join("dev/disk/by-id", "google-"+PD.PDName)
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		success, err := regexp.MatchString(deviceName, line)
		if err != nil {
			return -1, err
		}
		if success {
			refCount++
		}
	}
	return refCount, nil
}
