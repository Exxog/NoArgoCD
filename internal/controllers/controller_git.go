package controllers

import (
	"fmt"
	"time"

	"github.com/Exxog/NoArgoCD/internal/watchers"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ControllerGit gère les dépôts GitLab à surveiller
type ControllerGit struct {
	watcher *watchers.GitLabWatcher
	repos   []watchers.GitLabRepo
}

// NewControllerGit crée un nouveau contrôleur GitLab avec un watcher et un client
func NewControllerGit(client *gitlab.Client) *ControllerGit {
	controller := &ControllerGit{}
	controller.watcher = watchers.NewGitLabWatcher(controller, client)
	return controller
}

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerGit) AddRepository(url, branch string) {
	repo := watchers.GitLabRepo{URL: url, Branch: branch}
	c.watcher.AddRepository(repo)

}

// NotifyNewCommit est appelé par le watcher lorsqu'un nouveau commit est détecté
func (c *ControllerGit) NotifyNewCommit(repo watchers.GitLabRepo, commitID string) {
	fmt.Printf("📝 Nouveau commit sur %s [%s] : %s\n", repo.URL, repo.Branch, commitID)
}

// StartWatching démarre la surveillance des dépôts GitLab à intervalles réguliers
func (c *ControllerGit) StartWatching(interval time.Duration) {
	fmt.Println("🚀 Démarrage de la surveillance des dépôts GitLab...")
	c.watcher.Watch(interval)
}
