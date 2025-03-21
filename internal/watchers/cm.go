package watchers

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ConfigMapWatcher est responsable de la surveillance des ConfigMaps dans Kubernetes
type ConfigMapWatcher struct {
	clientset *kubernetes.Clientset
}

// NewConfigMapWatcher crée un nouveau ConfigMapWatcher
func NewConfigMapWatcher() (*ConfigMapWatcher, error) {
	k8sClient, err := setupKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &ConfigMapWatcher{
		clientset: k8sClient,
	}, nil
}

// setupKubernetesClient initialise le client Kubernetes à partir du kubeconfig
func setupKubernetesClient() (*kubernetes.Clientset, error) {
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
func containsKey(dataMap map[string]interface{}, key string) bool {
	_, exists := dataMap[key]
	return exists
}
func getFirstKey(dataMap map[string]interface{}) string {
	// On parcourt toutes les clés du dictionnaire
	for key := range dataMap {
		return key // On retourne la première clé trouvée
	}
	// Si aucune clé n'est trouvée, retourner une chaîne vide
	return ""
}

// Watch surveille les ConfigMaps dans un namespace donné
func (w *ConfigMapWatcher) Watch(namespace string, onUpdate func(*v1.ConfigMap)) {
	// Surveille les ConfigMaps dans le namespace spécifié
	fmt.Printf("🔍 Surveillance des ConfigMaps dans le namespace '%s'...\n", namespace)

	// Utilisation d'un Watcher Kubernetes pour surveiller les ConfigMaps
	// Ajout du contexte ici
	watchInterface, err := w.clientset.CoreV1().ConfigMaps(namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "nac=true",
	})
	if err != nil {
		log.Fatalf("❌ Erreur lors de la surveillance des ConfigMaps : %v", err)
	}

	// Le watcher observe les événements et appelle onUpdate chaque fois qu'un événement survient
	for event := range watchInterface.ResultChan() {
		// Correction ici : cast vers *v1.ConfigMap pour obtenir son nom
		fmt.Printf("Événement détecté : %v, ConfigMap: %s\n", event.Type, event.Object.(*v1.ConfigMap).Name)
		configMap := event.Object.(*v1.ConfigMap)
		switch event.Type {
		case "MODIFIED", "DELETED":
			fmt.Println("🛠️ Mise à jour détectée sur un ConfigMap : ", event.Type)
			// Ici, tu peux ajouter la logique pour extraire les informations des ConfigMaps et les envoyer à onUpdate
		case "ADDED":
			fmt.Println("🛠️ Mise à jour détectée sur un ConfigMap : ", event.Type)
			for key, value := range configMap.Data {
				fmt.Printf("Clé: %s, Valeur: %s\n", key, value)
				var dataMap map[string]interface{}

				// Désérialisation du YAML dans la map
				err := yaml.Unmarshal([]byte(value), &dataMap)
				if err != nil {
					fmt.Println("Erreur lors de la désérialisation de la clé", key, ":", err)
					continue
				}

				// Utilisation d'un switch pour vérifier la valeur de chaque clé
				switch getFirstKey(dataMap) {

				case "helm":
					fmt.Printf("➡️ La clé '%s' contient 'helm'.\n", key)
					// Traitement spécifique pour 'helm'
					fmt.Println("A") // Exemple de traitement pour 'helm'
					onUpdate(configMap)
				case "apply":
					fmt.Printf("➡️ La clé '%s' contient 'apply'.\n", key)
					// Traitement spécifique pour 'apply'
					fmt.Println("C") // Exemple de traitement pour 'apply'
				default:
					fmt.Printf("➡️ La clé '%s' ne contient ni 'helm', ni 'toto', ni 'apply'.\n", key)
				}
			}
		default:
			// Log pour afficher d'autres types d'événements qui pourraient se produire
			fmt.Println("Événement non traité:", event.Type)
		}
	}

	// Si jamais la surveillance finit sans rien détecter, on garde la goroutine active
	fmt.Println("Watcher terminé.")
	select {} // Pour éviter que la goroutine se termine immédiatement
}
