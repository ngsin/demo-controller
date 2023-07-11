package main

import (
	"fmt"
	"github.com/ngsin/demo-controller/pkg/apis/demoapi/v1alpha1"
	demoClients "github.com/ngsin/demo-controller/pkg/client/clientset/versioned"
	demoScheme "github.com/ngsin/demo-controller/pkg/client/clientset/versioned/scheme"
	demoAPIInformer "github.com/ngsin/demo-controller/pkg/client/informers/externalversions/demoapi/v1alpha1"
	demoAPILister "github.com/ngsin/demo-controller/pkg/client/listers/demoapi/v1alpha1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsInformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appsLister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

const controllerAgentName = "demo-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

type Controller struct {
	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// demoClientSet is a clientset for our own API group
	demoClientSet demoClients.Interface

	deploymentsLister appsLister.DeploymentLister
	demoAPIsLister    demoAPILister.DemoAPILister

	deploymentsSynced cache.InformerSynced
	demosSynced       cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(
	kubeClientSet kubernetes.Interface,
	demoClientSet demoClients.Interface,
	deploymentInformer appsInformers.DeploymentInformer,
	demoAPIInformer demoAPIInformer.DemoAPIInformer) *Controller {

	// Create event broadcaster
	// Add demo-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(demoScheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	// sent event to log
	eventBroadcaster.StartStructuredLogging(0)
	// sent to event sink，build event api resource
	eventBroadcaster.StartRecordingToSink(&typedCoreV1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, coreV1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientSet:     kubeClientSet,
		demoClientSet:     demoClientSet,
		deploymentsLister: deploymentInformer.Lister(),
		demoAPIsLister:    demoAPIInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		demosSynced:       demoAPIInformer.Informer().HasSynced,
		workqueue:         workqueue.NewRateLimitingQueueWithConfig(workqueue.DefaultControllerRateLimiter(), workqueue.RateLimitingQueueConfig{Name: "Demos"}),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Demo resources change
	demoAPIInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.Info("add demoAPI")
			controller.recorder.Event(obj.(runtime.Object), coreV1.EventTypeNormal, "demoAPI", "AddFunc")

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			klog.Info("update demoAPI")
			controller.recorder.Event(oldObj.(runtime.Object), coreV1.EventTypeNormal, "demoAPI", "AddFunc")
			klog.Infof("update demoAPI: %v", newObj.(*v1alpha1.DemoAPI).Spec.Image)
			time.Sleep(10 * time.Second)

		},
		DeleteFunc: func(obj interface{}) {
			klog.Info("delete demoAPI")
			controller.recorder.Event(obj.(runtime.Object), coreV1.EventTypeNormal, "demoAPI", "DeleteFunc")

		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.Info("add deployment")
			controller.recorder.Event(obj.(runtime.Object), coreV1.EventTypeNormal, "deployment", "AddFunc")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			klog.Info("update deployment")
			controller.recorder.Event(oldObj.(runtime.Object), coreV1.EventTypeNormal, "deployment", "UpdateFunc")
		},
		DeleteFunc: func(obj interface{}) {
			klog.Info("delete deployment")
			controller.recorder.Event(obj.(runtime.Object), coreV1.EventTypeNormal, "deployment", "DeleteFunc")

		},
	})

	return controller

}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Demo controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.demosSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	//Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	klog.Info("runWorker")
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
