package utils

import (
	"context"
	"fmt"
	"os"

	rolloutsclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var config *rest.Config
var deployRolloutDetails map[string]string
var clusterName string

// GetDeployments returns a list of all Kubernetes deployments in a namespace
func GetDeployments(namespace string) (map[string]string, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	// Get the deployments in the specified namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %v", err)
	}

	// Extract the deployment names
	for _, deployment := range deployments.Items {
		deployRolloutDetails[deployment.Name] = deployment.Spec.Template.Spec.Containers[0].Image
	}

	return deployRolloutDetails, nil
}

// GetArgoRollouts returns a list of all Argo Rollouts in a namespace
func GetArgoRollouts(namespace string) (map[string]string, error) {
	// Create a new rollouts clientset
	rolloutsClient, err := rolloutsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollouts clientset: %v", err)
	}

	// Get the rollouts in the specified namespace
	rollouts, err := rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get rollouts: %v", err)
	}

	// Extract the rollout names
	for _, rollout := range rollouts.Items {
		deployRolloutDetails[rollout.Name] = rollout.Spec.Template.Spec.Containers[0].Image
	}

	return deployRolloutDetails, nil
}

func GetAppDetails(namespace string) (map[string]string, error) {
	_, err := GetDeployments(namespace)
	if err != nil {
		return nil, err
	} else {
		return GetArgoRollouts(namespace)
	}

}

func init() {
	// Create a new clientset
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to local kubeconfig if in-cluster config fails
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = "/Users/ajames/.kube/config" // Default location for kubeconfig file
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			sugar.Errorf("failed to create config from kubeconfig: %v", err)
		}
	}

	deployRolloutDetails = make(map[string]string)
	if os.Getenv(CLUSTER_CONTEXT) != "" {
		clusterName = os.Getenv(CLUSTER_CONTEXT)
	} else {
		clusterName = "NOT_SET"
	}

}
