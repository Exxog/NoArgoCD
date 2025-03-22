package controllers

import (
	"encoding/json"
	"fmt"
	"time"

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

func installHelmChartFromGitOLD(repo watchers.GitLabRepo, chartPath, releaseName, namespace string) {
	namespace = utils.GetNamespace(namespace)
	println(repo.URL)
	utils.CloneOrUpdateRepo(repo.URL, "/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch), repo.Branch, "", "")
	utils.DeployOrUpdateHelmChartViaCmd("/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch)+"/"+chartPath, releaseName, namespace, "")
}

func installHelmChartFromGit(repo watchers.GitLabRepo, chartPath, releaseName, namespace, values string) {
	namespace = utils.GetNamespace(namespace)
	println(repo.URL)

	for {
		// Cloner ou mettre à jour le dépôt
		fmt.Println("[controller][helm] 📥 Clonage/Mise à jour du dépôt Git...")
		if err := utils.CloneOrUpdateRepo(repo.URL, "/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch), repo.Branch, "", ""); err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du clonage/mise à jour du dépôt: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		// Déployer ou mettre à jour le chart Helm
		fmt.Println("🚀 Déploiement du chart Helm...")
		err := utils.DeployOrUpdateHelmChartViaCmd("/tmp/"+utils.CleanFolderName(repo.URL+repo.Branch)+"/"+chartPath, releaseName, namespace, values)
		if err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du déploiement du chart: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		fmt.Println("[controller][helm] ✅ Déploiement réussi!")
		break // Sortir de la boucle si tout est OK
	}
}

func (c *ControllerHelm) InstallHelmChart(repo watchers.GitLabRepo) {
	helmCharts := getters.GetHelm(repo.URL, repo.Branch)

	for key, charts := range helmCharts {
		for _, chart := range charts {
			if repoURL, ok := chart["repoUrl"].(string); ok {
				fmt.Printf("[controller][helm] 🔹 Clé: %s, Repo URL: %s\n", key, repoURL)

				values := ""
				if _, exists := chart["values"]; exists {
					jsonString, _ := json.Marshal(chart["values"])
					values = string(jsonString)

				}

				installHelmChartFromGit(repo, chart["path"].(string), key, "", values)

			}
		}
	}

}

//faire un retry si helm marche pas ?

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerHelm) AddConfigMap(url, branch string) {
	fmt.Println("pass")
}
