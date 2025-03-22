package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
)

func WriteValuesFile(filePath, content string) error {
	if content == "" {
		return nil // Pas de valeurs à écrire, on passe
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors de la création du fichier values.yaml: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors de l'écriture dans values.yaml: %v", err)
	}

	fmt.Println("📄 Fichier values.yaml généré:", filePath)
	return nil
}

func DeployOrUpdateHelmChartViaCmdOLD(chartPath, releaseName, namespace string, valuesYaml string) error {
	// Tentative infinie
	for {
		// 1. Mettre à jour les dépendances du chart avec helm dependency update
		cmd := exec.Command("helm", "dependency", "update", chartPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println("[utils][helm] 📦 Mise à jour des dépendances...")
		if err := cmd.Run(); err != nil {
			fmt.Printf("[utils][helm] ❌ Erreur lors de la mise à jour des dépendances du chart: %v\n", err)
			// Attendre 30 secondes avant de réessayer
			fmt.Println("[utils][helm]⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		// 2. Déployer ou mettre à jour le chart avec helm upgrade --install
		fmt.Println("🚀 Déploiement ou mise à jour du chart...")
		upgradeCmd := exec.Command("helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--force", "-f", valuesYaml)
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr
		if err := upgradeCmd.Run(); err != nil {
			fmt.Printf("[utils][helm] ❌ Erreur lors de l'installation ou de la mise à jour du chart: %v\n", err)
			// Attendre 30 secondes avant de réessayer
			fmt.Println("[utils][helm] ⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		// Si tout s'est bien passé
		fmt.Println("✅ Déploiement réussi!")
		return nil
	}
}

func DeployOrUpdateHelmChartViaCmd(chartPath, releaseName, namespace, valuesYamlContent string) error {
	valuesFilePath := "/tmp/nac" + uuid.New().String() // Chemin temporaire pour le fichier YAML

	// Si des valeurs sont fournies, on les écrit dans values.yaml
	if err := WriteValuesFile(valuesFilePath, valuesYamlContent); err != nil {
		return err
	}

	// 1. Mettre à jour les dépendances du chart avec `helm dependency update`
	cmd := exec.Command("helm", "dependency", "update", chartPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("📦 Mise à jour des dépendances...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Erreur lors de la mise à jour des dépendances du chart: %v", err)
	}

	// 2. Construire la commande `helm upgrade --install`
	fmt.Println("🚀 Déploiement ou mise à jour du chart...")
	upgradeCmd := []string{
		"upgrade", "--install", releaseName, chartPath,
		"--namespace", namespace, "--force",
	}

	// Ajouter le fichier `values.yaml` seulement s'il a été généré
	if valuesYamlContent != "" {
		upgradeCmd = append(upgradeCmd, "-f", valuesFilePath)
	}
	fmt.Println(upgradeCmd)
	cmd = exec.Command("helm", upgradeCmd...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Erreur lors de l'installation ou de la mise à jour du chart: %v", err)
	}

	if valuesYamlContent != "" {
		if err := os.Remove(valuesFilePath); err != nil {
			fmt.Printf("⚠️ Impossible de supprimer le fichier temporaire %s: %v\n", valuesFilePath, err)
		} else {
			fmt.Println("🗑️ Fichier values.yaml supprimé:", valuesFilePath)
		}
	}
	fmt.Println("✅ Déploiement réussi!")
	return nil
}

func DeployOrUpdateHelmChartViaCmdOLD2(chartPath, releaseName, namespace string, valuesYaml string) error {
	// 1. Mettre à jour les dépendances du chart avec helm dependency update
	cmd := exec.Command("helm", "dependency", "update", chartPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("[utils][helm]📦 Mise à jour des dépendances...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[utils][helm] erreur lors de la mise à jour des dépendances du chart: %v", err)
	}

	// 2. Déployer ou mettre à jour le chart avec helm upgrade --install
	fmt.Println("🚀 Déploiement ou mise à jour du chart...")
	upgradeCmd := exec.Command("helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--force", "-f", valuesYaml)
	upgradeCmd.Stdout = os.Stdout
	upgradeCmd.Stderr = os.Stderr
	if err := upgradeCmd.Run(); err != nil {
		return fmt.Errorf("[utils][helm] erreur lors de l'installation ou de la mise à jour du chart: %v", err)
	}

	fmt.Println("✅ Déploiement réussi!")
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
	fmt.Println("🚀 Déploiement du chart...")
	chartSpec := &helmclient.ChartSpec{
		ReleaseName: releaseName,
		ChartName:   chartPath,
		Namespace:   namespace,
		ValuesYaml:  valuesYaml,
		UpgradeCRDs: true,
	}

	return client.InstallOrUpgradeChart(context.Background(), chartSpec, nil)
}
