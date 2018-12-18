package main

import (
	"flag"
	"github.com/appscode/go/signals"
	_"k8s.io/apimachinery/pkg/util/intstr"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
	"path/filepath"
	"time"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	_ "k8s.io/code-generator/_examples/crd/apis/example/v1"
	ctclientset "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	croncontrol "github.com/iamrz1/controller-for-custom-resource/controllers"
	informers "github.com/iamrz1/controller-for-custom-resource/pkg/client/informers/externalversions"
)
func main(){
	kubeFlag := flag.String("kubeconfig",filepath.Join(homedir.HomeDir(),".kube","config"),"Path to kubeconfig")
	flag.Parse()
	stopCh := signals.SetupSignalHandler()
	config , err := clientcmd.BuildConfigFromFlags("",*kubeFlag)
	if err != nil {
		panic(err)
	}
	cs, err := ctclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}


	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

//Create controller
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientSet, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(cs, time.Second*30)

	controller := croncontrol.NewController(clientSet, cs,
		kubeInformerFactory.Apps().V1().Deployments(),
		exampleInformerFactory.Examplecrd().V1().CronTabs())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	exampleInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
