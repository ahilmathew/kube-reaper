package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	expiryAnnotationKey = "kube-reaper/expires"
)

// Decide whether the tool is running locally or in a kube cluster
func main() {
	isLocal := getEnv("LOCAL")
	if isLocal {
		fmt.Printf("Using local credentials. Unset LOCAL or set LOCAL=true if you running from a Cluster\n")
		local()
	} else {
		fmt.Printf("Using inCluster credentials. Set LOCAL=true if you running locally\n")
		inCluster()
	}
}

// This will be used when the code is executed from your local machine that has access to the kubeconfig
func local() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	performNamespaceOperations(clientset)
}

// This will be used when the code is executed from a pod inside a kubernetes cluster
func inCluster() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	performNamespaceOperations(clientset)
}

func listNamespaces(clientset *kubernetes.Clientset) v1.NamespaceList {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list namespaces: %v\n", err)
		os.Exit(1)
	}
	return *namespaces
}

func performNamespaceOperations(clientset *kubernetes.Clientset) {
	// get a list of all namespaces in the cluster
	namespaces := listNamespaces(clientset)
	fmt.Printf("There are %d namespaces in the cluster\n", len(namespaces.Items))
	for _, ns := range namespaces.Items {
		expiryTimeStr := getExpiryTime(ns)
		if len(expiryTimeStr) > 0 {
			expiryTime, err := time.Parse("2006-01-02T15:04", getExpiryTime(ns))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse expiry time annotation %q for namespace %q: %v\n", expiryTimeStr, ns.Name, err)
				continue
			}
			if time.Now().After(expiryTime) {
				fmt.Printf("Deleting expired namespace %q\n", ns.Name)
				err = clientset.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to delete namespace %q: %v\n", ns.Name, err)
					continue
				}
				fmt.Printf("Deleted expired namespace %q\n", ns.Name)
			} else if time.Now().Before(expiryTime) {
				fmt.Printf("Skipping namespace %q as expiry time is in the future. ", ns.Name)
				ttl := time.Until(expiryTime)
				loc, _ := time.LoadLocation("Local")
				fmt.Printf("%q has %q until expiry (%q)\n", ns.Name, ttl, expiryTime.In(loc))
			}
		} else {
			fmt.Printf("Skipping namespace %q as no expiry time found\n", ns.Name)
		}
	}
}

func getExpiryTime(element v1.Namespace) string {
	return element.Annotations[expiryAnnotationKey]
}

func getEnv(v string) bool {
	ge := os.Getenv(v)
	if len(ge) > 0 {
		val, err := strconv.ParseBool(ge)
		if err != nil {
			fmt.Println("The environment variable value should be true or false.")
			panic(err.Error())
		}
		return val
	}
	return false
}

func checkPrefixes(n string) bool {
	allowedPrefixes := []string{"pr-", "-pr-", "-pr"}
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(n, prefix) {
			return true
		}
		if strings.HasSuffix(n, prefix) {
			return true
		}
		if strings.Contains(n, prefix) {
			return true
		}
	}
	return false
}
