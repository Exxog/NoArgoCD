package controllers

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/utils"
	"github.com/Exxog/NoArgoCD/internal/watchers"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// ControllerKube gère la surveillance des ConfigMaps et notifie ControllerGit
type ControllerKube struct {
	helmController *ControllerHelm
	watcher        *watchers.ConfigMapWatcher
}

// NewControllerKube crée un contrôleur pour surveiller les ConfigMaps et met à jour ControllerGit
func NewControllerKube(helmController *ControllerHelm) (*ControllerKube, error) {
	// Créer un ConfigMapWatcher
	ConfigMapWatcher, err := watchers.NewConfigMapWatcher()
	if err != nil {
		return nil, err
	}

	// Retourne une instance de ControllerKube
	return &ControllerKube{
		helmController: helmController,
		watcher:        ConfigMapWatcher,
	}, nil
}

// StartWatcher démarre la surveillance des ConfigMaps dans un namespace
func (c *ControllerKube) StartWatching(namespace string) {
	namespace = utils.GetNamespace(namespace)
	// Lancer la surveillance dans une nouvelle goroutine
	fmt.Println("[controllers][kube]🔍 Démarrage de la surveillance des ConfigMaps dans le namespace :", namespace)
	c.watcher.Watch(namespace, c.onUpdate, c.onDelete)
}

func getFirstKey(dataMap map[string]interface{}) string {
	// On parcourt toutes les clés du dictionnaire
	for key := range dataMap {
		return key // On retourne la première clé trouvée
	}
	// Si aucune clé n'est trouvée, retourner une chaîne vide
	return ""
}

// Callback exécutée lors d'une mise à jour de ConfigMap
func (c *ControllerKube) onUpdate(cm *v1.ConfigMap) {
	fmt.Println("[controllers][kube]⚡ Mise à jour des dépôts GitLab à partir des ConfigMaps Kubernetes !")
	//TODO filter sur le helm pour diriger vers le bon controller
	//c.helmController.AddCM(cm)
	for key, value := range cm.Data {
		//fmt.Printf("Clé: %s, Valeur: %s\n", key, value)
		var dataMap map[string]interface{}

		// Désérialisation du YAML dans la map
		err := yaml.Unmarshal([]byte(value), &dataMap)
		if err != nil {
			fmt.Println("[controllers][kube]Erreur lors de la désérialisation de la clé", key, ":", err)
			continue
		}

		// Utilisation d'un switch pour vérifier la valeur de chaque clé
		switch getFirstKey(dataMap) {

		case "helm":
			fmt.Printf("[controllers][kube]➡ CM contient clé '%s' contient 'helm'.\n", key)
			// Traitement spécifique pour 'helm'
			c.helmController.DetectHelmChartFromCM(dataMap, key)
		case "apply":
			fmt.Printf("[controllers][kube]➡️ CM contient '%s' contient 'helm'.\n", key)
			// Traitement spécifique pour 'helm'

		default:
			fmt.Printf("[controllers][kube]➡️ CM ne contient pas '%s' ne contient ni 'helm', ni 'toto', ni 'apply'.\n", key)
		}
	}
}
func (c *ControllerKube) onDelete(cm *v1.ConfigMap) {
	fmt.Println("[controllers][kube]⚡ Mise à jour des dépôts GitLab à partir des ConfigMaps Kubernetes !")
	//TODO filter sur le helm pour diriger vers le bon controller
	//c.helmController.AddCM(cm)
	for key, value := range cm.Data {
		//fmt.Printf("Clé: %s, Valeur: %s\n", key, value)
		var dataMap map[string]interface{}

		// Désérialisation du YAML dans la map
		err := yaml.Unmarshal([]byte(value), &dataMap)
		if err != nil {
			fmt.Println("[controllers][kube]Erreur lors de la désérialisation de la clé", key, ":", err)
			continue
		}

		// Utilisation d'un switch pour vérifier la valeur de chaque clé
		switch getFirstKey(dataMap) {

		case "helm":
			fmt.Printf("[controllers][kube]➡ CM contient clé '%s' contient 'helm'.\n", key)
			// Traitement spécifique pour 'helm'
			c.helmController.DeleteHelmChartFromCM(dataMap, key)
		case "apply":
			fmt.Printf("[controllers][kube]➡️ CM contient '%s' contient 'helm'.\n", key)
			// Traitement spécifique pour 'helm'

		default:
			fmt.Printf("[controllers][kube]➡️ CM ne contient pas '%s' ne contient ni 'helm', ni 'toto', ni 'apply'.\n", key)
		}
	}
}

// Fonction pour tester directement le ControllerKube sans passer par main.go
