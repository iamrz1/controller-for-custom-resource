package custom_clients

import (
	"fmt"
	ct "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"time"
)

func UpdateClient(cs *ct.Clientset)  {
	time.Sleep(time.Second*15)
	fmt.Println("Updating CronTab.")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		cronClient := cs.ExamplecrdV1().CronTabs(metav1.NamespaceDefault)
		result, getErr := cronClient.Get("my-cron-tab", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = 4                          // reduce replica count
		//result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := cronClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")

}
