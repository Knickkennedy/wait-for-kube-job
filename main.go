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
	if err != nil {
		panic(err.Error())
	}

	job, jobExists := os.LookupEnv("JOB_NAME")
	namespace, namespaceExists := os.LookupEnv("NAMESPACE")

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if !jobExists || job == "" {
		panic("JOB_NAME is a required environment variable.")
	} else if !namespaceExists || namespace == "" {
		panic("NAMESPACE is a required environment variable.")
	} else {
		for {
			fmt.Printf("Checking on job %s\n", job)
			done, err := GetJobStatus(job, namespace, clientset)

			time.Sleep(10 * time.Second)

			if err != nil {
				panic(err)
			} else if done {
				break
			}
		}
	}
}

func GetJobStatus(job, namespace string, clientset *kubernetes.Clientset) (bool, error) {

	result, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), job, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return true, fmt.Errorf("job %s in namespace %s not found\n", job, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return true, fmt.Errorf("Error getting job %s in namespace %s: %v\n",
			job, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found job %s in namespace %s\n", job, namespace)

		if result.Status.Failed > 0 {
			return true, fmt.Errorf("prequisite job %s in namespace %s failed. Please try again\n", job, namespace)
		} else if result.Status.Active == 0 && result.Status.Succeeded == 0 {
			fmt.Printf("Job %s hasn't started yet.\n", result.Name)
			return false, nil
		} else if result.Status.Active > 0 {
			fmt.Printf("Job %s is still running.\n", result.Name)
			return false, nil
		} else {
			fmt.Printf("Job %s has completed\n", result.Name)
			return true, nil
		}
	}
}
