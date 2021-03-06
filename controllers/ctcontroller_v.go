package controllers
import (
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	_"k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	crontabv1	"github.com/iamrz1/controller-for-custom-resource/pkg/apis/examplecrd.com/v1"
	crontabclientset "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	ctscheme "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned/scheme"
	crontabinformers "github.com/iamrz1/controller-for-custom-resource/pkg/client/informers/externalversions/examplecrd.com/v1"
	crontablisters "github.com/iamrz1/controller-for-custom-resource/pkg/client/listers/examplecrd.com/v1"

)

const controllerAgentName = "crontab-controller"

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// ctclientset is a clientset for our own API group
	ctclientset crontabclientset.Interface

	//deploymentsLister appslisters.DeploymentLister
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	ctLister        crontablisters.CronTabLister
	ctSynced        cache.InformerSynced
	//deploymentSynced cache.SharedIndexInformer

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

// NewController returns a new CronTab controller
func NewController(
	kubeclientsetP kubernetes.Interface,
	ctclientsetP crontabclientset.Interface,
	deploymentInformerP appsinformers.DeploymentInformer,
	ctInformerP crontabinformers.CronTabInformer) *Controller {
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(ctscheme.AddToScheme(scheme.Scheme))
	log.Println("Creating event broadcaster")

	// Create event broadcaster to setup recorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientsetP.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	//set controller struct variables
	controller := &Controller{
		kubeclientset:	kubeclientsetP,
		ctclientset:	ctclientsetP,
		deploymentsLister: deploymentInformerP.Lister(),
		deploymentsSynced: deploymentInformerP.Informer().HasSynced,
		ctLister:	ctInformerP.Lister(),
		ctSynced:	ctInformerP.Informer().HasSynced,
		workqueue:	workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CronTanWorkQueue"),
		recorder:	recorder,
	}

	log.Println("Setting up event handlers")
	// Set up an event handler for when CronTab resources change
	ctInformerP.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCronTab,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueCronTab(new)
		},
		
		//DeleteFunc: controller.enqueueCronTab,
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a crontab resource, will enqueue that resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformerP.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// enqueueCronTab takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueCronTab(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	log.Println("==========================================> handleObject is called.")
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "CronTab" {
			return
		}

		ctObject, err := c.ctLister.CronTabs(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueCronTab(ctObject)
		return
	}
}

// Run will set up the event handlers for types we are interested in, as well
// as syncronizing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Println("Starting CronTab controller")

	// Wait for the caches to be synced before starting workers
	log.Println("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.ctSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Println("Starting workers")
	// Launch two workers to process CronTab resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Println("Started workers")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}
// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}
// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	//Default obj = default/my-cron-tab
	fmt.Println("ProcessNextWorkItem==============>")
	fmt.Println(obj)
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
		key, ok = obj.(string)
		//Default => key =  default/my-cron-tab  Okay =  true

		if !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// CronTab resource to be synced.
		err := c.syncHandler(key)
		//default=> err = <nil>
		if err != nil && c.workqueue.NumRequeues(key)<5 {
			// Put the item back on the workqueue to handle any transient errors.
			fmt.Println("Adding o work queue")
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		fmt.Printf("Successfully synced '%s'", key)
		fmt.Println("")
		log.Printf("Successfully synced '%s'", key)
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
	res, err := c.ctLister.CronTabs(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := res.Spec.DeploymentName
	//deploymentName =  stupid-crontab-deployment-from-client-go
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		//res.Spec.DeploymentName = "default-crontab-deployment"
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Foo.spec
	deployment, err := c.deploymentsLister.Deployments(res.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {

		deployment, err = c.kubeclientset.AppsV1().Deployments(res.Namespace).Create(newDeployment(res))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and returns error message
	if !metav1.IsControlledBy(deployment, res) {
		msg := fmt.Sprintf("Message: Resource Exists = "+ deployment.Name)
		fmt.Println(">>>>>>>>>>>> 07", msg)
		c.recorder.Event(res, corev1.EventTypeWarning, "resourceExists", msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	fmt.Println(">>>>>>>>>>>> 08")
	if res.Spec.Replicas != 0 && res.Spec.Replicas != *deployment.Spec.Replicas {
		klog.V(4).Infof("Foo %s replicas: %d, deployment replicas: %d", name, res.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1().Deployments(res.Namespace).Update(newDeployment(res))
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	fmt.Println(">>>>>>>>>>>> 09")
	if err != nil {
		return err
	}
	fmt.Println(">>>>>>>>>>>> 10")

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = c.updateFooStatus(res, deployment)
	if err != nil {
		return err
	}

	c.recorder.Event(res, corev1.EventTypeNormal, "SuccessSynced", "MessageResourceSynced")
	return nil
}

func (c *Controller) updateFooStatus(ct *crontabv1.CronTab, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	ctCopy := ct.DeepCopy()
	ctCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.ctclientset.ExamplecrdV1().CronTabs(ct.Namespace).Update(ctCopy)
	return err
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(ct *crontabv1.CronTab) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "book-server",
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ct.Spec.DeploymentName,
			Namespace: ct.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ct, schema.GroupVersionKind{
					Group:   crontabv1.SchemeGroupVersion.Group,
					Version: crontabv1.SchemeGroupVersion.Version,
					Kind:    "CronTab",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &ct.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: ct.Spec.Template.Spec,
			},
		},
	}
}