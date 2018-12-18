## practice repo (client-go/CRD) 
Contains codes to implement calls to kube-api-server from go to practice CRUD functions on kubernetes/custom resources.

### Using client-go (or any custom client)
In go program, do:
1. Use filepath.join() to join and add [homedir](k8s.io/client-go/util/homedir)/.kube/config to flag. Parse flag.
2. Build congig from flags using [clientcmd](k8s.io/client-go/tools/clientcmd).
3. Get clientset for that config file from specific clientsets(.NewForConfig(config)).
4. Define object for that api.
5. Perform CRUD operation.

### Custom Resources in GO
First create a custom resource definition [file](/yaml/crontabClient.yaml) in yaml format.
Now an object of that type can me made using yaml [file](/yaml/crontabDefination.yaml)s.

To create a go client for custom object creation,  a few things are needed to be done.
1. A [code generator](https://github.com/kubernetes/code-generator) is needed to be cloned in vendor/k8s.io.
2. Create file [register.go](pkg/apis/examplecrd.com/register.go)  in pkg/apis/<group-name>/ directory.
3. Create files [doc.go](pkg/apis/examplecrd.com/v1/doc.go), [register.go](pkg/apis/examplecrd.com/v1/register.go),
and [types.go](pkg/apis/examplecrd.com/v1/types.go) in pkg/apis/<group-name>/<api-version>/ directory.
4. Now create a shell script [update-codegen.sh](hack/update-codegen.sh) in /hack directory.
run the script from root directory of the project:
    
    ``
    $hack/update-codegen.sh 
    ``
5. Now objects of newly created custom type can be created, and client APIs can be invoked from go program.

