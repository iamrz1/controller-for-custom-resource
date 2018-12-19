package custom_clients

import (
	"fmt"
	crontabv1 "github.com/iamrz1/controller-for-custom-resource/pkg/apis/examplecrd.com/v1"
	ct "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)
func RunClients(kubeFlag string){
	config , err := clientcmd.BuildConfigFromFlags("",kubeFlag)
	if err != nil {
		panic(err)
	}
	cs, err := ct.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cron := &crontabv1.CronTab{
		ObjectMeta: ObjectMeta{
			Name:"my-cron-tab",
			Namespace:NamespaceDefault,
			Labels: map[string]string{
					"run":"book-server-client",
			},
		},
		Spec: crontabv1.CronTabDeploymentSpec{
			DeploymentName:"stupid-crontab-deployment-from-client-go",
			Replicas: 2,
			Template:crontabv1.CronTabPodTemplate{
				ObjectMeta: ObjectMeta{
					Name:"cron-pod",
					Namespace:NamespaceDefault,
					Labels: map[string]string{
						"run":"book-server-client",
					},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:"book-server-with-client-go",
							Args:[]string{
								"-v","-b",
							},
							Image:"rezoan/api_server:1.0.1",
							ImagePullPolicy: "IfNotPresent",
							Ports:[]corev1.ContainerPort{
								{
									ContainerPort:8080,
									Protocol: corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	fmt.Println("Creating cronTab")
	newct, err := cs.ExamplecrdV1().CronTabs(NamespaceDefault).Create(cron)
	fmt.Println("cronTab created")
	fmt.Println("cronTab = ",newct)
	//time.Sleep(time.Second*15)
}

