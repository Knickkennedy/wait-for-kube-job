package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {

	// This section of code is only necessary if we are external to the cluster

	/*

		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}

		var job *string = flag.String("job", "", "The specific job you're looking for")
		flag.Parse()
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	*/

	// This is needed for internal to cluster authentication
	config, err := rest.InClusterConfig()

	job, exists := os.LookupEnv("JOB_NAME")

	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if exists {
		done := false
		for !done {
			/*jobs, err := clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("There are %d jobs in the cluster\n", len(jobs.Items))*/

			namespace := "db2"
			result, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), job, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("Job %s in namespace %s not found\n", job, namespace)
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("Error getting job %s in namespace %s: %v\n",
					job, namespace, statusError.ErrStatus.Message)
			} else if err != nil {
				panic(err.Error())
			} else {
				fmt.Printf("Found job %s in namespace %s\n", job, namespace)

				if result.Status.Active == 0 && result.Status.Succeeded == 0 && result.Status.Failed == 0 {
					fmt.Printf("Job %s hasn't started yet.\n", result.Name)
				} else if result.Status.Active > 0 {
					fmt.Printf("Job %s is still running.\n", result.Name)
				} else if result.Status.Succeeded > 0 {
					fmt.Printf("Job %s has completed\n", result.Name)
					done = true
					break
				}
			}
			time.Sleep(10 * time.Second)
			fmt.Printf("Checking on job %s\n", job)
		}
	} else {
		fmt.Println("Environment Variable \"JOB_NAME\" is required.")
	}
}
