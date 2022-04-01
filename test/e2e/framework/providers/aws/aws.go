/*
Copyright 2018 The Kubernetes Authors.

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

package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"

	v1 "k8s.io/api/core/v1"
<<<<<<< HEAD
	"k8s.io/kubernetes/test/e2e/framework"
=======
	"k8s.io/kubernetes/test/e2e/framework/config"
	"k8s.io/kubernetes/test/e2e/framework/providers"
>>>>>>> fcf4212fb3d (restructure e2e framework and its sub-packages to solve potential  cycle imports.)
	e2epv "k8s.io/kubernetes/test/e2e/framework/pv"
	"k8s.io/kubernetes/test/e2e/framework/utils"
	awscloud "k8s.io/legacy-cloud-providers/aws"
)

func init() {
	providers.RegisterProvider("aws", newProvider)
}

func newProvider() (providers.ProviderInterface, error) {
	if config.TestContext.CloudConfig.Zone == "" {
		utils.Logf("Warning: gce-zone not specified! Some tests that use the AWS SDK may select the wrong region and fail.")
	}
	return &Provider{}, nil
}

// Provider is a structure to handle AWS clouds for e2e testing
type Provider struct {
	providers.NullProvider
}

// ResizeGroup resizes an instance group
func (p *Provider) ResizeGroup(group string, size int32) error {
	awsSession, err := session.NewSession()
	if err != nil {
		return err
	}
	client := autoscaling.New(awsSession)
	return awscloud.ResizeInstanceGroup(client, group, int(size))
}

// GroupSize returns the size of an instance group
func (p *Provider) GroupSize(group string) (int, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return -1, err
	}
	client := autoscaling.New(awsSession)
	instanceGroup, err := awscloud.DescribeInstanceGroup(client, group)
	if err != nil {
		return -1, fmt.Errorf("error describing instance group: %v", err)
	}
	if instanceGroup == nil {
		return -1, fmt.Errorf("instance group not found: %s", group)
	}
	return instanceGroup.CurrentSize()
}

// DeleteNode deletes a node which is specified as the argument
func (p *Provider) DeleteNode(node *v1.Node) error {
	client := newAWSClient("")

	instanceID, err := awscloud.KubernetesInstanceID(node.Spec.ProviderID).MapToAWSInstanceID()
	if err != nil {
		return err
	}

	req := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(string(instanceID)),
		},
	}
	_, err = client.TerminateInstances(req)
	return err
}

func (p *Provider) CreateShare() (string, string, string, error) {
	return "", "", "", nil
}

func (p *Provider) DeleteShare(accountName, shareName string) error {
	return nil
}

// CreatePD creates a persistent volume on the specified availability zone
func (p *Provider) CreatePD(zone string) (string, error) {
	client := newAWSClient(zone)
	request := &ec2.CreateVolumeInput{}
	request.AvailabilityZone = aws.String(zone)
	request.Size = aws.Int64(10)
	request.VolumeType = aws.String(awscloud.DefaultVolumeType)

	// We need to tag the volume so that locked-down IAM configurations can still mount it
	if config.TestContext.CloudConfig.ClusterTag != "" {
		clusterID := config.TestContext.CloudConfig.ClusterTag

		legacyTag := &ec2.Tag{
			Key:   aws.String(awscloud.TagNameKubernetesClusterLegacy),
			Value: aws.String(clusterID),
		}

		newTag := &ec2.Tag{
			Key:   aws.String(awscloud.TagNameKubernetesClusterPrefix + clusterID),
			Value: aws.String(awscloud.ResourceLifecycleOwned),
		}

		tagSpecification := &ec2.TagSpecification{
			ResourceType: aws.String(ec2.ResourceTypeVolume),
			Tags:         []*ec2.Tag{legacyTag, newTag},
		}

		request.TagSpecifications = append(request.TagSpecifications, tagSpecification)
	}

	response, err := client.CreateVolume(request)
	if err != nil {
		return "", err
	}

	az := aws.StringValue(response.AvailabilityZone)
	awsID := aws.StringValue(response.VolumeId)

	volumeName := "aws://" + az + "/" + awsID
	return volumeName, nil
}

// DeletePD deletes a persistent volume
func (p *Provider) DeletePD(pdName string) error {
	client := newAWSClient("")

	tokens := strings.Split(pdName, "/")
	awsVolumeID := tokens[len(tokens)-1]

	request := &ec2.DeleteVolumeInput{VolumeId: aws.String(awsVolumeID)}
	_, err := client.DeleteVolume(request)
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok && awsError.Code() == "InvalidVolume.NotFound" {
			utils.Logf("volume deletion implicitly succeeded because volume %q does not exist.", pdName)
		} else {
			return fmt.Errorf("error deleting EBS volumes: %v", err)
		}
	}
	return nil
}

// CreatePVSource creates a persistent volume source
func (p *Provider) CreatePVSource(zone, diskName string) (*v1.PersistentVolumeSource, error) {
	return &v1.PersistentVolumeSource{
		AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
			VolumeID: diskName,
			FSType:   "ext3",
		},
	}, nil
}

// DeletePVSource deletes a persistent volume source
func (p *Provider) DeletePVSource(pvSource *v1.PersistentVolumeSource) error {
	return e2epv.DeletePDWithRetry(pvSource.AWSElasticBlockStore.VolumeID)
}

func newAWSClient(zone string) *ec2.EC2 {
	var cfg *aws.Config

	if zone == "" {
		zone = config.TestContext.CloudConfig.Zone
	}
	if zone == "" {
		utils.Logf("Warning: No AWS zone configured!")
		cfg = nil
	} else {
		region := zone[:len(zone)-1]
		cfg = &aws.Config{Region: aws.String(region)}
	}
	session, err := session.NewSession()
	if err != nil {
		utils.Logf("Warning: failed to create aws session")
	}
	return ec2.New(session, cfg)
}
