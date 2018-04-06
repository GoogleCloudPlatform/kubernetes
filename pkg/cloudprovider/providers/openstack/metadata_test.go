/*
Copyright 2016 The Kubernetes Authors.

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

package openstack

import (
	"strings"
	"testing"
)

var FakeMetadata = Metadata{
	UUID:             "83679162-1378-4288-a2d4-70e13ec132aa",
	Hostname:         "test",
	AvailabilityZone: "nova",
}

func SetMetadataFixture(value *Metadata) {
	metadataCache = value
}

func ClearMetadata() {
	metadataCache = nil
}

func TestParseMetadata(t *testing.T) {
	_, err := parseMetadata(strings.NewReader("bogus"))
	if err == nil {
		t.Errorf("Should fail when bad data is provided: %s", err)
	}

	data := strings.NewReader(`
{
    "availability_zone": "nova",
    "files": [
        {
            "content_path": "/content/0000",
            "path": "/etc/network/interfaces"
        },
        {
            "content_path": "/content/0001",
            "path": "known_hosts"
        }
    ],
    "hostname": "test.novalocal",
    "launch_index": 0,
    "name": "test",
    "meta": {
        "role": "webservers",
        "essential": "false"
    },
    "public_keys": {
        "mykey": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDBqUfVvCSez0/Wfpd8dLLgZXV9GtXQ7hnMN+Z0OWQUyebVEHey1CXuin0uY1cAJMhUq8j98SiW+cU0sU4J3x5l2+xi1bodDm1BtFWVeLIOQINpfV1n8fKjHB+ynPpe1F6tMDvrFGUlJs44t30BrujMXBe8Rq44cCk6wqyjATA3rQ== Generated by Nova\n"
    },
    "uuid": "83679162-1378-4288-a2d4-70e13ec132aa",
    "devices": [
        {
            "bus": "scsi",
            "serial": "6df1888b-f373-41cf-b960-3786e60a28ef",
            "tags": ["fake_tag"],
            "type": "disk",
            "address": "0:0:0:0"
        }
    ]
}
`)
	md, err := parseMetadata(data)
	if err != nil {
		t.Fatalf("Should succeed when provided with valid data: %s", err)
	}

	if md.Hostname != "test" {
		t.Errorf("incorrect hostname: %s", md.Name)
	}

	if md.UUID != "83679162-1378-4288-a2d4-70e13ec132aa" {
		t.Errorf("incorrect uuid: %s", md.UUID)
	}

	if md.AvailabilityZone != "nova" {
		t.Errorf("incorrect az: %s", md.AvailabilityZone)
	}

	if len(md.Devices) != 1 {
		t.Errorf("expecting to find 1 device, found %d", len(md.Devices))
	}

	if md.Devices[0].Bus != "scsi" {
		t.Errorf("incorrect disk bus: %s", md.Devices[0].Bus)
	}

	if md.Devices[0].Address != "0:0:0:0" {
		t.Errorf("incorrect disk address: %s", md.Devices[0].Address)
	}

	if md.Devices[0].Type != "disk" {
		t.Errorf("incorrect device type: %s", md.Devices[0].Type)
	}

	if md.Devices[0].Serial != "6df1888b-f373-41cf-b960-3786e60a28ef" {
		t.Errorf("incorrect device serial: %s", md.Devices[0].Serial)
	}
}
