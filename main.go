package main

import (
	"flag"
	demoClients "github.com/ngsin/demo-controller/pkg/client/clientset/versioned"
	demoInformers "github.com/ngsin/demo-controller/pkg/client/informers/externalversions"
	"github.com/ngsin/demo-controller/pkg/signals"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"time"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {

	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	demoClient, err := demoClients.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building demo clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeInformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	demoInformerFactory := demoInformers.NewSharedInformerFactory(demoClient, time.Second*30)

	controller := NewController(kubeClient, demoClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		demoInformerFactory.Demogroup().V1alpha1().DemoAPIs(),
	)

	kubeInformerFactory.Start(stopCh)
	demoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
