# Controller for custom resource CronTab (client-go/CRD) 
Contains codes to implement controller to control custom resource CronTab that deploys kube-api-server from client-go and yaml files.

## Usage
1. Create custom resource definition using [crontabDefination.yaml](yaml/crontabDefination.yaml) 
2. Run the program. Controller for CronTab resources will be created. That will be followed by the creation of a crontab resource.
3. CronTab resource creates a deployment with 2 replica pods.   
4. Stop the program, and delete the CronTab resource.`kubectl delete ct --all`