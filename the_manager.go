package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	director director.Director
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
	return &BoshInstances{Cloud: b}, false
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

func buildUAA(cfg BCCMConfig) (boshuaa.UAA, error) {
	logger := boshlog.NewLogger(boshlog.LevelError)
	factory := boshuaa.NewFactory(logger)
	config, err := boshuaa.NewConfigFromURL(fmt.Sprintf("https://%s:8443", cfg.Host))
	if err != nil {
		return nil, err
	}
	config.Client = cfg.Client
	config.ClientSecret = cfg.ClientSecret
	config.CACert = cfg.CACert
	return factory.New(config)
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
	uaa, err := buildUAA(cfg)
	if err != nil {
		glog.Fatalf("Coudn't create UAA client: %s", err.Error())
		os.Exit(1)
	}
	factorConfig := director.FactoryConfig{
		Host:         cfg.Host,
		Port:         25555,
		Client:       cfg.Client,
		ClientSecret: cfg.ClientSecret,
		CACert:       cfg.CACert,
		TokenFunc:    boshuaa.NewClientTokenSession(uaa).TokenFunc,
	}
	directorFactory.New(factorConfig, nil, nil)

	b := BCCM{}
	return &b, nil
}

func init() {
	cloudprovider.RegisterCloudProvider("BOSH", BCCMFactory)
}
