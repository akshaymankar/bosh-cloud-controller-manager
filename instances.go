package main

import (
	"context"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type BoshInstances struct {
	Cloud *BCCM
}

func (i *BoshInstances) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	return []v1.NodeAddress{}, nil
}

func (i *BoshInstances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	instanceGroup, instanceUUID := parseProviderID(providerID)
	deps, err := i.Cloud.director.Deployments()
	if err != nil {
		return []v1.NodeAddress{}, nil
	}

	dep := deps[0]
	vms, err := dep.VMInfos()
	if err != nil {
		return []v1.NodeAddress{}, nil
	}

	for _, vm := range vms {
		if vm.JobName == instanceGroup && vm.ID == instanceUUID {
			glog.V(1).Infof("Found Instance: %s, ID: %s in Deployment: %s", instanceGroup, instanceUUID, dep.Name())
		}
	}

	return []v1.NodeAddress{{Type: v1.NodeExternalDNS, Address: "some-address"}}, nil
}

func parseProviderID(providerID string) (string, string) {
	splits := strings.Split(providerID, "/")
	return splits[0], splits[1]
}

func (i *BoshInstances) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return "", nil
}

func (i *BoshInstances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return "", nil
}

func (i *BoshInstances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	return "", nil
}

func (i *BoshInstances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	return "", nil
}

func (i *BoshInstances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return nil
}

func (i *BoshInstances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName("some-node"), nil
}

func (i *BoshInstances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	return false, nil
}
