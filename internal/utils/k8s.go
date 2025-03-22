package utils

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func SetupKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Si un kubeconfig est présent dans l'environnement
	if home := homedir.HomeDir(); home != "" {
		kubeconfig := fmt.Sprintf("%s/.kube/config", home)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("erreur de chargement du kubeconfig: %v", err)
		}
	} else {
		// Utilisation de la configuration par défaut si aucun kubeconfig trouvé
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("erreur de connexion au cluster Kubernetes: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erreur de création du client Kubernetes: %v", err)
	}

	return clientset, nil
}

func GetConfigMapsWithLabel(clientset *kubernetes.Clientset, namespace, label string) ([]v1.ConfigMap, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: label,
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur lors de la récupération des ConfigMaps : %v", err)
	}

	return configMaps.Items, nil
}

func GetNamespace(namespace string) string {
	// Si le namespace est vide, tente de lire le fichier du namespace
	if namespace == "" {
		data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			// Si la lecture échoue, utiliser "default" comme namespace
			fmt.Println("⚠️ Erreur lors de la lecture du namespace du fichier, utilisation du namespace par défaut.")
			return "default"
		}
		// Si la lecture est réussie, retourner la valeur lue dans le fichier
		return string(data)
	}
	// Si le namespace n'est pas vide, retourner sa valeur
	return namespace
}
