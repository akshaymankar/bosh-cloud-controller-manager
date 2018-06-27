package main

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type BoshInstances struct{}

func (i *BoshInstances) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	return []v1.NodeAddress{}, nil
}

func (i *BoshInstances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	return []v1.NodeAddress{}, nil
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
