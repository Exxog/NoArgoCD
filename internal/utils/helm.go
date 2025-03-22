package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
)

func DeployOrUpdateHelmChartViaCmd(chartPath, releaseName, namespace string, valuesYaml string) error {
	// Tentative infinie
	for {
		// 1. Mettre à jour les dépendances du chart avec helm dependency update
		cmd := exec.Command("helm", "dependency", "update", chartPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println("📦 Mise à jour des dépendances...")
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Erreur lors de la mise à jour des dépendances du chart: %v\n", err)
			// Attendre 30 secondes avant de réessayer
			fmt.Println("⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		// 2. Déployer ou mettre à jour le chart avec helm upgrade --install
		fmt.Println("🚀 Déploiement ou mise à jour du chart...")
		upgradeCmd := exec.Command("helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--force", "-f", valuesYaml)
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr
		if err := upgradeCmd.Run(); err != nil {
			fmt.Printf("❌ Erreur lors de l'installation ou de la mise à jour du chart: %v\n", err)
			// Attendre 30 secondes avant de réessayer
			fmt.Println("⏳ Tentative après 30 secondes...")
			time.Sleep(30 * time.Second)
			continue // Réessayer
		}

		// Si tout s'est bien passé
		fmt.Println("✅ Déploiement réussi!")
		return nil
	}
}

func DeployOrUpdateHelmChartViaCmdOLD(chartPath, releaseName, namespace string, valuesYaml string) error {
	// 1. Mettre à jour les dépendances du chart avec helm dependency update
	cmd := exec.Command("helm", "dependency", "update", chartPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("📦 Mise à jour des dépendances...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erreur lors de la mise à jour des dépendances du chart: %v", err)
	}

	// 2. Déployer ou mettre à jour le chart avec helm upgrade --install
	fmt.Println("🚀 Déploiement ou mise à jour du chart...")
	upgradeCmd := exec.Command("helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--force", "-f", valuesYaml)
	upgradeCmd.Stdout = os.Stdout
	upgradeCmd.Stderr = os.Stderr
	if err := upgradeCmd.Run(); err != nil {
		return fmt.Errorf("erreur lors de l'installation ou de la mise à jour du chart: %v", err)
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
		return nil, fmt.Errorf("erreur lors de l'initialisation du client Helm: %v", err)
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
