package controllers

import (
	"fmt"
	"time"

	"github.com/Exxog/NoArgoCD/internal/watchers"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Controller gère les dépôts surveillés
type Controller struct {
	watcher *watchers.GitLabWatcher
}

// NewController crée un nouveau contrôleur avec un watcher
func NewController(client *gitlab.Client) *Controller {
	controller := &Controller{}
	controller.watcher = watchers.NewGitLabWatcher(controller, client)
	return controller
}

// AddRepository ajoute un dépôt à surveiller
func (c *Controller) AddRepository(url, branch string) {
	repo := watchers.GitLabRepo{URL: url, Branch: branch}
	c.watcher.AddRepository(repo)
}

// NotifyNewCommit est appelé par le watcher lorsqu'un commit est détecté
func (c *Controller) NotifyNewCommit(repo watchers.GitLabRepo, commitID string) {
	fmt.Printf("📝 Nouveau commit sur %s [%s] : %s\n", repo.URL, repo.Branch, commitID)
}

// StartWatching démarre la surveillance des dépôts
func (c *Controller) StartWatching(interval time.Duration) {
	fmt.Println("🚀 Démarrage de la surveillance des dépôts GitLab...")
	c.watcher.Watch(interval)
}
