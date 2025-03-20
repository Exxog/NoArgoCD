package controllers

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/watchers"
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
	gitController *ControllerGit
	watcher       *watchers.ConfigMapWatcher
}

// NewControllerKube crée un contrôleur pour surveiller les ConfigMaps et met à jour ControllerGit
func NewControllerKube(gitController *ControllerGit) (*ControllerKube, error) {
	// Créer un ConfigMapWatcher
	ConfigMapWatcher, err := watchers.NewConfigMapWatcher()
	if err != nil {
		return nil, err
	}

	// Retourne une instance de ControllerKube
	return &ControllerKube{
		gitController: gitController,
		watcher:       ConfigMapWatcher,
	}, nil
}

// StartWatcher démarre la surveillance des ConfigMaps dans un namespace
func (c *ControllerKube) StartWatcher(namespace string) {
	// Lancer la surveillance dans une nouvelle goroutine
	go func() {
		fmt.Println("🎯 Démarrage de la surveillance des ConfigMaps dans le namespace :", namespace)
		c.watcher.Watch(namespace, c.onConfigMapUpdate)
	}()
}

// Callback exécutée lors d'une mise à jour de ConfigMap
func (c *ControllerKube) onConfigMapUpdate(newRepos []watchers.GitLabRepo) {
	fmt.Println("⚡ Mise à jour des dépôts GitLab à partir des ConfigMaps Kubernetes !")
	// Le ControllerGit reçoit la mise à jour des dépôts à surveiller
	c.gitController.UpdateRepos(newRepos)
}

// Fonction pour tester directement le ControllerKube sans passer par main.go
