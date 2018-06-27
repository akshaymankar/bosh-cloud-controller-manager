package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-cli/director"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

type BCCMConfig struct {
	Host         string `yaml:"bosh-environment"`
	CACert       string `yaml:"bosh-ca-cert"`
	Client       string `yaml:"bosh-client"`
	ClientSecret string `yaml:"bosh-client-secret"`
}

type BCCM struct {
	director   director.Director
	kubeclient kubernetes.Interface
}

func (b *BCCM) Initialize(clientBuilder controller.ControllerClientBuilder) {
	kube, err := clientBuilder.Client("foo")
	if err != nil {
		glog.Error("Cannot get kubeclient")
		os.Exit(1)
	}

	b.kubeclient = kube
}

func (b *BCCM) untaint() {
	watcher, err := b.kubeclient.CoreV1().Nodes().Watch(metav1.ListOptions{})
	if err != nil {
		glog.Error("Cannot watch nodes!")
		os.Exit(1)
	}
	defer watcher.Stop()
	nodesChan := watcher.ResultChan()
	for nodeEvent := range nodesChan {
		glog.V(1).Infof("Event of type %s occurred", nodeEvent.Type)
		node := nodeEvent.Object.(*v1.Node)
		glog.V(1).Infof("Node Name: %s", node.Name)
		glog.V(1).Infof("Node Addresses: %#v", node.Status.Addresses)
		glog.V(1).Infof("Node: %#v", node)
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
			b.kubeclient.CoreV1().Nodes().Update(node)
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
	glog.V(1).Info("Giving Instances!!!!!")
	return &BoshInstances{Cloud: b}, true
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
	c, err := ioutil.ReadAll(config)
	if err != nil {
		glog.Fatalf("Coudn't read the config with error %s", err.Error())
		os.Exit(1)
	}
	cfg := BCCMConfig{}
	yaml.Unmarshal(c, &cfg)

	directorFactory := director.NewFactory(boshlog.NewLogger(boshlog.LevelDebug))
	fc, err := director.NewConfigFromURL(cfg.Host)
	fc.Client = cfg.Client
	fc.ClientSecret = cfg.ClientSecret
	fc.CACert = cfg.CACert
	if err != nil {
		glog.Fatalf("Coudn't read the config with error %s", err.Error())
		os.Exit(1)
	}
	d, err := directorFactory.New(fc, nil, nil)
	if err != nil {
		glog.Fatalf("Coudn't read the config with error %s", err.Error())
		os.Exit(1)
	}

	glog.V(1).Infof("Factory Config: %#v", fc)

	b := BCCM{director: d}
	return &b, nil
}

func init() {
	cloudprovider.RegisterCloudProvider("BOSH", BCCMFactory)
}
