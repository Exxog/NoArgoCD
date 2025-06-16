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
	controller := &ControllerHelm{
		gitController: gitController,
	}
	controller.watcher = watchers.NewHelmWatcher(controller)
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
	authSecretName, _ := helmData["authSecretName"].(string)
	//values, _ := helmData["values"].(string)
	values := utils.ConvertToYaml(helmData)

	//chartPath := helm["helm"].(map[interface{}]interface{})["path"]
	//values := helm["helm"].(map[interface{}]interface{})["values"]
	//values = utils.ConvertToYaml(helm["helm"].(map[interface{}]interface{}))

	fmt.Println("DETECTION!!! ", helm)
	c.gitController.AddRepository(watchers.GitRepo{URL: repoURL, Branch: targetRevision, AuthSecretName: authSecretName})
	installHelmChartFromGit(watchers.GitRepo{URL: repoURL, Branch: targetRevision, AuthSecretName: authSecretName}, chartPath, releaseName, config.Namespace, values)
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
		namespace = config.Namespace
	}

	c.Remove(namespace, repoURL, targetRevision, releaseName)

}
func (c *ControllerHelm) Remove(namespace, repoURL, targetRevision, releaseName string) {
	if len(getters.GetHelm(repoURL, targetRevision, namespace)) == 0 {
		removeCacheHelm(releaseName)
		c.gitController.RemoveRepository(repoURL, targetRevision)
	}
	utils.DeleteHelmRelease(releaseName, namespace)
}

func removeCacheHelm(releaseName string) {
	//TODOdelete cacheRepo et dependances
	cachePath := filepath.Join(os.Getenv("HOME"), ".cache", "helm", "archives", releaseName+"-*.tgz")

	// Trouver et supprimer les fichiers
	files, _ := filepath.Glob(cachePath)
	for _, file := range files {
		os.Remove(file)
	}
}

func installHelmChartFromGit(repo watchers.GitRepo, chartPath, releaseName, namespace string, values []byte) {
	namespace = utils.GetNamespace(namespace)
	println(repo.URL)

	for {
		// Cloner ou mettre à jour le dépôt
		fmt.Println("[controller][helm] 📥 Clonage/Mise à jour du dépôt Git...")
		username, password, _ := utils.GetUsernamePasswordFromSecret(config.Namespace, repo.AuthSecretName)
		if err := utils.CloneOrUpdateRepo(repo.URL, utils.DestClonePath(repo.URL, repo.Branch), repo.Branch, username, password); err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du clonage/mise à jour du dépôt: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			//time.Sleep(30 * time.Second)
			//continue // Réessayer
		}

		// Déployer ou mettre à jour le chart Helm
		fmt.Println("[controller][helm]🚀 Déploiement du chart Helm...")
		err := utils.DeployOrUpdateHelmChartViaCmd(utils.DestClonePath(repo.URL, repo.Branch)+"/"+chartPath, releaseName, namespace, values)
		if err != nil {
			fmt.Printf("[controller][helm] ❌ Erreur lors du déploiement du chart: %v\n", err)
			fmt.Println("[controller][helm] ⏳ Tentative après 30 secondes...")
			//time.Sleep(30 * time.Second)
			//continue // Réessayer
		}

		fmt.Println("[controller][helm] \033[32mDéploiement réussi!\033[0m")
		break
	}
}

func (c *ControllerHelm) InstallHelmChart(repo watchers.GitRepo) {

	helmCharts := getters.GetHelm(repo.URL, repo.Branch, config.Namespace)

	for key, charts := range helmCharts {
		for _, chart := range charts {
			switch {
			case chart["repoUrl"] != nil:
				// Chart depuis un repo Git
				if repoURL, ok := chart["repoUrl"].(string); ok {
					fmt.Printf("[controller][helm] 🔹 Clé: %s, Repo URL: %s\n", key, repoURL)
					yamlData := utils.ConvertToYaml(chart)
					installHelmChartFromGit(repo, chart["path"].(string), key, config.Namespace, yamlData)
				}
			case chart["oci"] != nil:
				// Chart depuis un repo Helm distant (OCI ou repo add)
				if chartName, ok := chart["oci"].(string); ok {
					fmt.Printf("[controller][helm] 🔹 Clé: %s, oci: %s\n", key, chartName)
					yamlData := utils.ConvertToYaml(chart)
					// Appelle ici ta fonction d'installation pour les charts distants
					utils.InstallHelmChartFromOCI(chartName, key, config.Namespace, yamlData)
				}
			default:
				fmt.Printf("[controller][helm] ⚠️ Clé %s : type de chart non reconnu\n", key)
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
