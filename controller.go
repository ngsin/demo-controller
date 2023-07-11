package main

import (
	"context"
	"fmt"
	"github.com/ngsin/demo-controller/pkg/apis/demoapi/v1alpha1"
	demoClients "github.com/ngsin/demo-controller/pkg/client/clientset/versioned"
	demoScheme "github.com/ngsin/demo-controller/pkg/client/clientset/versioned/scheme"
	demoAPIInformer "github.com/ngsin/demo-controller/pkg/client/informers/externalversions/demoapi/v1alpha1"
	demoAPILister "github.com/ngsin/demo-controller/pkg/client/listers/demoapi/v1alpha1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		AddFunc: controller.enqueueDemo,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueueDemo(newObj)

		},
		DeleteFunc: controller.enqueueDemo,
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
	for c.processNextWorkItem() {
	}
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueDemo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true

}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	demo, err := c.demoAPIsLister.DemoAPIs(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := demo.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Foo.spec
	deployment, err := c.deploymentsLister.Deployments(demo.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeClientSet.AppsV1().Deployments(demo.Namespace).Create(context.TODO(), newDeployment(demo), metaV1.CreateOptions{})
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and return error msg.
	if !metaV1.IsControlledBy(deployment, demo) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		c.recorder.Event(demo, coreV1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if demo.Spec.Replicas != nil && *demo.Spec.Replicas != *deployment.Spec.Replicas {
		klog.V(4).Infof("Foo %s replicas: %d, deployment replicas: %d", name, *demo.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeClientSet.AppsV1().Deployments(demo.Namespace).Update(context.TODO(), newDeployment(demo), metaV1.UpdateOptions{})
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = c.updateDemoStatus(demo, deployment)
	if err != nil {
		return err
	}

	c.recorder.Event(demo, coreV1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateDemoStatus(demo *v1alpha1.DemoAPI, deployment *appsV1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	demoCopy := demo.DeepCopy()
	demoCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.demoClientSet.DemogroupV1alpha1().DemoAPIs(demo.Namespace).UpdateStatus(context.TODO(), demoCopy, metaV1.UpdateOptions{})
	return err
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(demo *v1alpha1.DemoAPI) *appsV1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": demo.Name,
	}

	return &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      demo.Spec.DeploymentName,
			Namespace: demo.Namespace,
			OwnerReferences: []metaV1.OwnerReference{
				*metaV1.NewControllerRef(demo, v1alpha1.SchemeGroupVersion.WithKind("DemoAPI")),
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: demo.Spec.Replicas,
			Selector: &metaV1.LabelSelector{
				MatchLabels: labels,
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: labels,
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  "demo-container",
							Image: demo.Spec.Image,
						},
					},
				},
			},
		},
	}
}
