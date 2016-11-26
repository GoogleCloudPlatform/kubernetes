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

package route53

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"strings"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordSets = ResourceRecordSets{}

type ResourceRecordSets struct {
	zone *Zone
}

func (rrsets ResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	input := route53.ListResourceRecordSetsInput{
		HostedZoneId: rrsets.zone.impl.Id,
	}

	var list []dnsprovider.ResourceRecordSet
	err := rrsets.zone.zones.interface_.service.ListResourceRecordSetsPages(&input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		for _, rrset := range page.ResourceRecordSets {
			list = append(list, &ResourceRecordSet{rrset, &rrsets})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (rrsets ResourceRecordSets) Get(name string) (dnsprovider.ResourceRecordSet, error) {
	input := route53.ListResourceRecordSetsInput{
		HostedZoneId:    rrsets.zone.impl.Id,
		StartRecordName: aws.String(name),
		MaxItems:        aws.String("2"), // get 2 items to make sure there is no duplicate records with same DNS name
	}

	var list []dnsprovider.ResourceRecordSet
	err := rrsets.zone.zones.interface_.service.ListResourceRecordSetsPages(&input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		for _, rrset := range page.ResourceRecordSets {
			if strings.TrimSuffix(name, ".") == strings.TrimSuffix(aws.StringValue(rrset.Name), ".") {
				list = append(list, &ResourceRecordSet{rrset, &rrsets})
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (r ResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &ResourceRecordChangeset{
		zone:   r.zone,
		rrsets: &r,
	}
}

func (r ResourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	rrstypeStr := string(rrstype)
	rrs := &route53.ResourceRecordSet{
		Name: &name,
		Type: &rrstypeStr,
		TTL:  &ttl,
	}
	for _, rrdata := range rrdatas {
		rrs.ResourceRecords = append(rrs.ResourceRecords, &route53.ResourceRecord{
			Value: aws.String(rrdata),
		})
	}

	return ResourceRecordSet{
		rrs,
		&r,
	}
}
