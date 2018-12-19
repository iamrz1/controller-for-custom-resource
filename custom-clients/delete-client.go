package custom_clients

import (
	"fmt"
	ct "github.com/iamrz1/controller-for-custom-resource/pkg/client/clientset/versioned"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

)

func DeleteClient(cs *ct.Clientset ){
	time.Sleep(time.Second*120)
	fmt.Println("Deleting cronTab")
	err := cs.ExamplecrdV1().CronTabs(NamespaceDefault).Delete("my-cron-tab",NewDeleteOptions(0))
	if err != nil {
		panic(err)
	}
	fmt.Println("cronTab Deleted")

	//time.Sleep(time.Second*15)
}

