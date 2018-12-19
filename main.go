package main

import (
	"flag"
	"fmt"
	"github.com/appscode/go/signals"
	croncontrol "github.com/iamrz1/controller-for-custom-resource/controllers"
	"github.com/iamrz1/controller-for-custom-resource/custom-clients"
	ctclientset "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	informers "github.com/iamrz1/controller-for-custom-resource/pkg/client/informers/externalversions"
	_ "k8s.io/apimachinery/pkg/util/intstr"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	_ "k8s.io/code-generator/_examples/crd/apis/example/v1"
	"k8s.io/klog"
	"path/filepath"
	"time"
)
func main(){
	kubeFlag := flag.String("kubeconfig",filepath.Join(homedir.HomeDir(),".kube","config"),"Path to kubeconfig")
	flag.Parse()
	config , err := clientcmd.BuildConfigFromFlags("",*kubeFlag)
	if err != nil {
		panic(err)
	}
	//Setup clientSets for resources
	cs, err := ctclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//Setup for controller
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientSet, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(cs, time.Second*30)

	fmt.Println("Instantiate controller.")
	//instantiate controller
	controller := croncontrol.NewController(clientSet, cs,
		kubeInformerFactory.Apps().V1().Deployments(),
		exampleInformerFactory.Examplecrd().V1().CronTabs())

	fmt.Println("Start InformerFactory")
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	stopCh := signals.SetupSignalHandler()
	kubeInformerFactory.Start(stopCh)
	exampleInformerFactory.Start(stopCh)

	fmt.Println("call to client-go goes here")
	go custom_clients.CreateClient(cs)
	go custom_clients.UpdateClient(cs)
	go custom_clients.DeleteClient(cs)

	//Run controller
	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
	fmt.Println("=====================>6")

}
