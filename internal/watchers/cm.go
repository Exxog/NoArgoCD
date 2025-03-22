package watchers

import (
	"context"
	"fmt"
	"log"

	"github.com/Exxog/NoArgoCD/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapWatcher est responsable de la surveillance des ConfigMaps dans Kubernetes
type ConfigMapWatcher struct {
	clientset *kubernetes.Clientset
}

// NewConfigMapWatcher crée un nouveau ConfigMapWatcher
func NewConfigMapWatcher() (*ConfigMapWatcher, error) {
	k8sClient, err := utils.SetupKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &ConfigMapWatcher{
		clientset: k8sClient,
	}, nil
}

// setupKubernetesClient initialise le client Kubernetes à partir du kubeconfig

func containsKey(dataMap map[string]interface{}, key string) bool {
	_, exists := dataMap[key]
	return exists
}

// Watch surveille les ConfigMaps dans un namespace donné
func (w *ConfigMapWatcher) Watch(namespace string, onUpdate func(*v1.ConfigMap)) {
	for {
		// Surveille les ConfigMaps dans le namespace spécifié
		fmt.Printf("[watchers][cm]🔍 Surveillance des ConfigMaps dans le namespace '%s'...\n", namespace)

		// Utilisation d'un Watcher Kubernetes pour surveiller les ConfigMaps
		// Ajout du contexte ici
		watchInterface, err := w.clientset.CoreV1().ConfigMaps(namespace).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: "nac=true",
		})
		if err != nil {
			log.Fatalf("[watchers][cm]❌ Erreur lors de la surveillance des ConfigMaps : %v", err)
		}

		// Le watcher observe les événements et appelle onUpdate chaque fois qu'un événement survient
		for event := range watchInterface.ResultChan() {
			// Correction ici : cast vers *v1.ConfigMap pour obtenir son nom
			fmt.Printf("[watchers][cm]📝 Événement détecté : %v, ConfigMap: %s\n", event.Type, event.Object.(*v1.ConfigMap).Name)
			configMap := event.Object.(*v1.ConfigMap)
			switch event.Type {
			case "MODIFIED", "DELETED":
				fmt.Println("[watchers][cm]🛠📝️ Mise à jour détectée sur un ConfigMap : ", event.Type)
				// Ici, tu peux ajouter la logique pour extraire les informations des ConfigMaps et les envoyer à onUpdate
			case "ADDED":
				fmt.Println("[watchers][cm]🛠️📝 Mise à jour détectée sur un ConfigMap : ", event.Type)
				onUpdate(configMap)

			default:
				// Log pour afficher d'autres types d'événements qui pourraient se produire
				fmt.Println("[watchers][cm]Événement non traité:", event.Type)
			}
		}

		// Si jamais la surveillance finit sans rien détecter, on garde la goroutine active
	}
}
