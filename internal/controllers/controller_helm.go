package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Exxog/NoArgoCD/internal/config"
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
	repos         []watchers.GitRepo
	watcher       *watchers.HelmWatcher
}

// NewControllerGit crée un nouveau contrôleur GitLab avec un watcher et un client
func NewControllerHelm(gitController *ControllerGit) *ControllerHelm {

	HelmWatcher := watchers.NewHelmWatcher()

	controller := &ControllerHelm{
		watcher:       HelmWatcher,
		gitController: gitController,
	}
	return controller
}
func (c *ControllerHelm) DetectHelmChartFromCM(helm map[string]any, releaseName string) {
	fmt.Println("[controllers][helm]🔄 ADD HELM")
	fmt.Println(helm["helm"])
	helmData, ok := helm["helm"].(map[interface{}]interface{})

	if !ok {
		fmt.Println("❌ Erreur de récupération des données du chart.")
		return
	}

	repoURL, _ := helmData["repoUrl"].(string)
	targetRevision, _ := helmData["targetRevision"].(string)
	chartPath, _ := helmData["path"].(string)
	//values, _ := helmData["values"].(string)
	values := utils.ConvertToYaml(helmData)

	//chartPath := helm["helm"].(map[interface{}]interface{})["path"]
	//values := helm["helm"].(map[interface{}]interface{})["values"]
	//values = utils.ConvertToYaml(helm["helm"].(map[interface{}]interface{}))

	fmt.Println("DETECTION!!! ", helm)
	c.gitController.AddRepository(repoURL, targetRevision)
	installHelmChartFromGit(watchers.GitRepo{URL: repoURL, Branch: targetRevision}, chartPath, releaseName, "", values)
}

func (c *ControllerHelm) DeleteHelmChartFromCM(helm map[string]any, releaseName string) {
	helmData, ok := helm["helm"].(map[interface{}]interface{})
	if !ok {
		fmt.Println("❌ Erreur de récupération des données du chart.")
		return
	}

	repoURL, repoURLOk := helmData["repoUrl"].(string)
	targetRevision, revOk := helmData["targetRevision"].(string)

	if !repoURLOk || !revOk {
		fmt.Println("❌ repoUrl ou targetRevision manquants.")
		return
	}

	fmt.Println("[controllers][helm] DELETE HELM")

	namespace, namespaceOk := helmData["namespace"].(string)
	if !namespaceOk {
		namespace = ""
	}

	if len(getters.GetHelm(repoURL, targetRevision, namespace)) == 0 {
		//TODOdelete cacheRepo et dependances
		cachePath := filepath.Join(os.Getenv("HOME"), ".cache", "helm", "archives", releaseName+"-*.tgz")

		// Trouver et supprimer les fichiers
		files, _ := filepath.Glob(cachePath)
		for _, file := range files {
			os.Remove(file)
		}

		c.gitController.RemoveRepository(repoURL, targetRevision)
	}
	utils.DeleteHelmRelease(releaseName, namespace)
}

func installHelmChartFromGit(repo watchers.GitRepo, chartPath, releaseName, namespace string, values []byte) {
	namespace = utils.GetNamespace(namespace)
	println(repo.URL)

	for {
		// Cloner ou mettre à jour le dépôt
		fmt.Println("[controller][helm] 📥 Clonage/Mise à jour du dépôt Git...")
		if err := utils.CloneOrUpdateRepo(repo.URL, config.NacTmpDir+utils.CleanFolderName(repo.URL+repo.Branch), repo.Branch, "", ""); err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du clonage/mise à jour du dépôt: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			//time.Sleep(30 * time.Second)
			//continue // Réessayer
		}

		// Déployer ou mettre à jour le chart Helm
		fmt.Println("[controller][helm]🚀 Déploiement du chart Helm...")
		err := utils.DeployOrUpdateHelmChartViaCmd(config.NacTmpDir+utils.CleanFolderName(repo.URL+repo.Branch)+"/"+chartPath, releaseName, namespace, values)
		if err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du déploiement du chart: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			//time.Sleep(30 * time.Second)
			//continue // Réessayer
		}

		fmt.Println("[controller][helm] ✅ Déploiement réussi!")
		break // Sortir de la boucle si tout est OK
	}
}

func (c *ControllerHelm) InstallHelmChart(repo watchers.GitRepo) {

	helmCharts := getters.GetHelm(repo.URL, repo.Branch, "")

	for key, charts := range helmCharts {
		for _, chart := range charts {
			if repoURL, ok := chart["repoUrl"].(string); ok {

				fmt.Printf("[controller][helm] 🔹 Clé: %s, Repo URL: %s\n", key, repoURL)
				yamlData := utils.ConvertToYaml(chart)

				installHelmChartFromGit(repo, chart["path"].(string), key, "", yamlData)

			}
		}
	}

}

//faire un retry si helm marche pas ?

// AddRepository ajoute un dépôt GitLab à surveiller
func (c *ControllerHelm) AddConfigMap(url, branch string) {
	fmt.Println("pass")
}

// StartWatching démarre la surveillance des dépôts GitLab à intervalles réguliers
func (c *ControllerHelm) StartWatching(interval time.Duration) {
	fmt.Println("[controllers][helm]🔄🌐🗂️ Démarrage de la surveillance des Helms")
	c.watcher.Watch(interval)
}
