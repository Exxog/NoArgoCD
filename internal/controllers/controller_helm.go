package controllers

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/watchers"
	v1 "k8s.io/api/core/v1"
)

// ControllerGit gère les dépôts GitLab à surveiller
type ControllerHelm struct {
	gitController *ControllerGit
	repos         []watchers.GitLabRepo
}

func (c *ControllerHelm) Add(cm *v1.ConfigMap) {
	fmt.Println("🔄 ADD CM")
	c.gitController.AddRepository("https://gitlab.com/gitlab-org/gitlab", "master")

}

// NewControllerGit crée un nouveau contrôleur GitLab avec un watcher et un client
func NewControllerHelm(gitController *ControllerGit) *ControllerHelm {
	controller := &ControllerHelm{
		gitController: gitController,
	}
	return controller
}

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerHelm) AddConfigMap(url, branch string) {
	fmt.Println("pass")
}
