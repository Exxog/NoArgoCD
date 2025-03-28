package controllers

import (
	"fmt"
	"time"

	"github.com/Exxog/NoArgoCD/internal/utils"
	"github.com/Exxog/NoArgoCD/internal/watchers"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ControllerGit gère les dépôts GitLab à surveiller
type ControllerGit struct {
	helmController *ControllerHelm
	watcher        *watchers.GitWatcher
	repos          []watchers.GitRepo
}

// NewControllerGit crée un nouveau contrôleur GitLab avec un watcher et un client
func NewControllerGit(client *gitlab.Client, helmController *ControllerHelm) *ControllerGit {
	controller := &ControllerGit{
		helmController: helmController,
	}
	controller.watcher = watchers.NewGitWatcher(controller)
	return controller
}

func (c *ControllerGit) SetHelmController(helmController *ControllerHelm) {
	c.helmController = helmController
}

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerGit) AddRepository(url, branch string) {
	repo := watchers.GitRepo{URL: url, Branch: branch}
	c.watcher.AddRepository(repo)

}
func (c *ControllerGit) RemoveRepository(url, branch string) {
	repo := watchers.GitRepo{URL: url, Branch: branch}
	c.watcher.RemoveRepository(repo)

}

// NotifyNewCommit est appelé par le watcher lorsqu'un nouveau commit est détecté
func (c *ControllerGit) NotifyNewCommit(repo watchers.GitRepo, commitID string) {
	fmt.Printf("[controllers][git] ✨🌐🗂️  Nouveau commit sur %s [%s] : %s\n", repo.URL, repo.Branch, commitID)
	utils.CloneOrUpdateRepo(repo.URL, "/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch), repo.Branch, "", "")
	c.helmController.InstallHelmChart(repo)

}

// StartWatching démarre la surveillance des dépôts GitLab à intervalles réguliers
func (c *ControllerGit) StartWatching(interval time.Duration) {
	fmt.Println("[controllers][git]🔄🌐🗂️ Démarrage de la surveillance des dépôts GitLab...")
	c.watcher.Watch(interval)
}

// UpdateRepos met à jour les repos surveillés dans ControllerGit
func (c *ControllerGit) UpdateRepos(repos []watchers.GitRepo) {
	fmt.Println("[controllers][kube] 🔄 Mise à jour des dépôts GitLab à surveiller")
	c.repos = repos
	// Ici, tu peux lancer le watcher GitLab pour surveiller les nouveaux repos
	// c.startWatching() - Exemple, si tu as une méthode pour commencer à surveiller les repos
}
