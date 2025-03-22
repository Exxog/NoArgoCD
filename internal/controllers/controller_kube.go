package controllers

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/utils"
	"github.com/Exxog/NoArgoCD/internal/watchers"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// UpdateRepos met à jour les repos surveillés dans ControllerGit
func (c *ControllerGit) UpdateRepos(repos []watchers.GitLabRepo) {
	fmt.Println("🔄 Mise à jour des dépôts GitLab à surveiller")
	c.repos = repos
	// Ici, tu peux lancer le watcher GitLab pour surveiller les nouveaux repos
	// c.startWatching() - Exemple, si tu as une méthode pour commencer à surveiller les repos
}

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
	fmt.Println("🔍 Démarrage de la surveillance des ConfigMaps dans le namespace :", namespace)
	c.watcher.Watch(namespace, c.onConfigMapUpdate)
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
func (c *ControllerKube) onConfigMapUpdate(cm *v1.ConfigMap) {
	fmt.Println("⚡ Mise à jour des dépôts GitLab à partir des ConfigMaps Kubernetes !")
	//TODO filter sur le helm pour diriger vers le bon controller
	//c.helmController.AddCM(cm)
	for key, value := range cm.Data {
		//fmt.Printf("Clé: %s, Valeur: %s\n", key, value)
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
			c.helmController.DetectHelmChartFromCM(dataMap)
		case "apply":
			fmt.Printf("➡️ La clé '%s' contient 'helm'.\n", key)
			// Traitement spécifique pour 'helm'
			fmt.Println("A") // Exemple de traitement pour 'helm'

		default:
			fmt.Printf("➡️ La clé '%s' ne contient ni 'helm', ni 'toto', ni 'apply'.\n", key)
		}
	}
}

// Fonction pour tester directement le ControllerKube sans passer par main.go
