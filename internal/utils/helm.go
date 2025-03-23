package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
)

func DeployOrUpdateHelmChartViaCmd(chartPath, releaseName, namespace string, valuesYamlContent []byte) error {
	valuesFilePath := "/tmp/nac" + uuid.New().String() // Chemin temporaire pour le fichier YAML
	// Si des valeurs sont fournies, on les écrit dans values.yaml
	if err := WriteYAMLToFile(valuesFilePath, valuesYamlContent); err != nil {
		return err
	}

	// 1. Mettre à jour les dépendances du chart avec `helm dependency update`
	cmd := exec.Command("helm", "dependency", "update", chartPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("[utils][helm] 📦 Mise à jour des dépendances...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[utils][helm]❌ Erreur lors de la mise à jour des dépendances du chart: %v", err)
	}

	// 2. Construire la commande `helm upgrade --install`
	fmt.Println("[utils][helm]🚀 Déploiement ou mise à jour du chart...")
	upgradeCmd := []string{
		"upgrade", "--install", releaseName, chartPath,
		"--namespace", namespace, "--force", "--set", "metadata.labels.nac='true'",
	}

	// Ajouter le fichier `values.yaml` seulement s'il a été généré
	if len(valuesYamlContent) != 0 {
		upgradeCmd = append(upgradeCmd, "-f", valuesFilePath)
	}
	fmt.Println(upgradeCmd)
	cmd = exec.Command("helm", upgradeCmd...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	//if len(valuesYamlContent) != 0 {
	//	if err := os.Remove(valuesFilePath); err != nil {
	//		fmt.Printf("[utils][helm]⚠️ Impossible de supprimer le fichier temporaire %s: %v\n", valuesFilePath, err)
	//	} else {
	//		fmt.Println("[utils][helm] 🗑️ Fichier values.yaml supprimé:", valuesFilePath)
	//	}
	//}

	if err != nil {
		return fmt.Errorf("[utils][helm]❌ Erreur lors de l'installation ou de la mise à jour du chart: %v", err)
	}

	fmt.Println("[utils][helm] ✅ Déploiement réussi!")
	return nil
}

//"--recreate-pods",

// DeployOrUpdateHelmChart installe ou met à jour un chart Helm local
func DeployOrUpdateHelmChart(chartPath, releaseName, namespace string, valuesYaml string) (*release.Release, error) {
	// Initialisation du client Helm
	opt := &helmclient.Options{
		Namespace:        namespace,
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
	}
	client, err := helmclient.New(opt)
	if err != nil {
		return nil, fmt.Errorf("[utils][helm] erreur lors de l'initialisation du client Helm: %v", err)
	}

	// Installation ou mise à jour du chart
	fmt.Println("[utils][helm]🚀 Déploiement du chart...")
	chartSpec := &helmclient.ChartSpec{
		ReleaseName: releaseName,
		ChartName:   chartPath,
		Namespace:   namespace,
		ValuesYaml:  valuesYaml,
		UpgradeCRDs: true,
	}

	return client.InstallOrUpgradeChart(context.Background(), chartSpec, nil)
}

func DeleteHelmRelease(releaseName, namespace string) error {
	// 1. Construire la commande `helm uninstall`
	cmd := exec.Command("helm", "uninstall", releaseName, "--namespace", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("[utils][helm] 🗑️ Suppression de la release %s dans le namespace %s...\n", releaseName, namespace)

	// Exécuter la commande
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("[utils][helm]❌ Erreur lors de la suppression de la release %s: %v", releaseName, err)
	}

	fmt.Printf("[utils][helm] ✅ La release %s a été supprimée avec succès du namespace %s.\n", releaseName, namespace)
	return nil
}

func GetHelmReleases(namespace string) ([]string, error) {
	// Construire la commande Helm list
	cmd := exec.Command("helm", "list", "--namespace", namespace, "-q")

	// Exécuter la commande et récupérer la sortie
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur lors de l'exécution de 'helm list': %v", err)
	}

	// Transformer la sortie en une liste de noms de release
	releases := strings.Split(strings.TrimSpace(out.String()), "\n")

	// Si aucune release trouvée, retourner une liste vide
	if len(releases) == 1 && releases[0] == "" {
		return []string{}, nil
	}

	return releases, nil
}

type HelmRelease struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

func GetHelmReleasesFiltered(namespace, labelKey, labelValue string) ([]string, error) {
	// Construire la commande Helm list avec sortie JSON
	cmd := exec.Command("helm", "list", "--namespace", namespace, "--output", "json")

	// Exécuter la commande et récupérer la sortie
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur lors de l'exécution de 'helm list': %v", err)
	}

	// Transformer la sortie JSON en une liste de HelmRelease
	var releases []HelmRelease
	if err := json.Unmarshal(out.Bytes(), &releases); err != nil {
		return nil, fmt.Errorf("❌ Erreur de parsing JSON : %v", err)
	}

	// Filtrer les releases par label
	var filteredReleases []string
	for _, release := range releases {
		if val, ok := release.Labels[labelKey]; ok && val == labelValue {
			filteredReleases = append(filteredReleases, release.Name)
		}
	}

	return filteredReleases, nil
}
