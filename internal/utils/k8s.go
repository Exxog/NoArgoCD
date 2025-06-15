package utils

import (
	"context"
	"fmt"
	"log"
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

func GetAuthSecret(namespace string, secretName string) (map[string][]byte, error) {
	clientset, err := SetupKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur lors de la configuration du client Kubernetes: %v", err)
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur lors de la récupération du secret %s dans le namespace %s: %v", secretName, namespace, err)
	}

	return secret.Data, nil
}

func GetUsernamePasswordFromSecret(namespace, secretName string) (string, string, error) {
	secretData, err := GetAuthSecret(namespace, secretName)
	if err != nil {
		return "", "", err
	}

	username := ""
	password := ""

	if val, ok := secretData["username"]; ok {
		username = string(val)
		log.Printf("Username found in secret: %s\n", username)
	}
	if val, ok := secretData["password"]; ok {
		password = string(val)
		log.Printf("Password found in secret: %s\n", password)
	}

	return username, password, nil
}
