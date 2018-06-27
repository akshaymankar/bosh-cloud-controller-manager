package main

import (
	"io"
	"os"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

type BCCM struct {
}

func (b *BCCM) Initialize(clientBuilder controller.ControllerClientBuilder) {
	kube, err := clientBuilder.Client("foo")
	if err != nil {
		glog.Error("Cannot get kubeclient")
		os.Exit(1)
	}
	watcher, err := kube.CoreV1().Nodes().Watch(metav1.ListOptions{})
	if err != nil {
		glog.Error("Cannot watch nodes!")
		os.Exit(1)
	}
	defer watcher.Stop()
	nodesChan := watcher.ResultChan()
	for nodeEvent := range nodesChan {
		glog.V(1).Infof("Event of type %s occurred.\nEvent Object: %#v", nodeEvent.Type, nodeEvent)
		node := nodeEvent.Object.(*v1.Node)
		var taintIndex int
		taintIndex = -1
		for i, taint := range node.Spec.Taints {
			if taint.Key == algorithm.TaintExternalCloudProvider {
				taintIndex = i
				break
			}
		}
		if taintIndex != -1 {
			node.Spec.Taints = remove(node.Spec.Taints, taintIndex)
			kube.CoreV1().Nodes().Update(node)
		}
	}
}

func remove(s []v1.Taint, i int) []v1.Taint {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (b *BCCM) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return nil, false
}

func (b *BCCM) Instances() (cloudprovider.Instances, bool) {
	return &BoshInstances{}, false
}

func (b *BCCM) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (b *BCCM) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (b *BCCM) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (b *BCCM) ProviderName() string {
	return "BOSH"
}

func (b *BCCM) HasClusterID() bool {
	return true
}

func BCCMFactory(config io.Reader) (cloudprovider.Interface, error) {
	return &BCCM{}, nil
}

func init() {
	cloudprovider.RegisterCloudProvider("BOSH", BCCMFactory)
}
