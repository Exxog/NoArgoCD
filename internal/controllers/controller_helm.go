package controllers

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/getters"
	"github.com/Exxog/NoArgoCD/internal/utils"
	"github.com/Exxog/NoArgoCD/internal/watchers"
)

type chartHelm struct {
	repoUrl    string
	path       string
	name       string
	chart      string
	revision   string
	authSecret struct {
	}
	values map[string]interface{}
}

// ControllerGit gère les dépôts GitLab à surveiller
type ControllerHelm struct {
	gitController *ControllerGit
	repos         []watchers.GitLabRepo
}

// NewControllerGit crée un nouveau contrôleur GitLab avec un watcher et un client
func NewControllerHelm(gitController *ControllerGit) *ControllerHelm {
	controller := &ControllerHelm{
		gitController: gitController,
	}
	return controller
}

func (c *ControllerHelm) DetectHelmChartFromCM(helm map[string]any) {
	fmt.Println("🔄 ADD HELM")
	fmt.Println(helm["helm"])
	repoURL := helm["helm"].(map[interface{}]interface{})["repoUrl"]
	targetRevision := helm["helm"].(map[interface{}]interface{})["targetRevision"]

	c.gitController.AddRepository(repoURL.(string), targetRevision.(string))

}

func installHelmChartFromGit(repo watchers.GitLabRepo, chartPath, releaseName, namespace string) {
	namespace = utils.GetNamespace(namespace)
	println(repo.URL)
	utils.CloneOrUpdateRepo(repo.URL, "/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch), repo.Branch, "", "")
	utils.DeployOrUpdateHelmChartViaCmd("/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch)+"/"+chartPath, releaseName, namespace, "")
}

func (c *ControllerHelm) InstallHelmChart(repo watchers.GitLabRepo) {
	helmCharts := getters.GetHelm(repo.URL, repo.Branch)

	for key, charts := range helmCharts {
		for _, chart := range charts {
			if repoURL, ok := chart["repoUrl"].(string); ok {
				fmt.Printf("🔹 Clé: %s, Repo URL: %s\n", key, repoURL)
				installHelmChartFromGit(repo, chart["path"].(string), key, "")

			}
		}
	}

}

//faire un retry si helm marche pas ?

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerHelm) AddConfigMap(url, branch string) {
	fmt.Println("pass")
}
