package watchers

import (
	"log"
	"time"

	"github.com/Exxog/NoArgoCD/internal/config"
	"github.com/Exxog/NoArgoCD/internal/getters"
	"github.com/Exxog/NoArgoCD/internal/utils"
)

// HelmWatcher surveille une release Helm à intervalle régulier
type HelmWatcher struct {
	controller  string
	namespace   string
	releaseName string
}

// NewHelmWatcher crée une nouvelle instance de HelmWatcher
func NewHelmWatcher() *HelmWatcher {
	return &HelmWatcher{
		releaseName: "",
		namespace:   config.Namespace,
	}
}

func GetHelmWithoutCM(keys, helm []string) []string {
	// Créer une map pour les clés pour une recherche rapide
	keyMap := make(map[string]struct{})
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}

	// Chercher les Helm releases qui ne sont pas présentes dans les keys
	var missingHelm []string
	for _, helmRelease := range helm {
		if _, exists := keyMap[helmRelease]; !exists {
			missingHelm = append(missingHelm, helmRelease)
		}
	}

	return missingHelm
}

func (w *HelmWatcher) Watch(interval time.Duration) {
	log.Printf("[watchers][helm] Surveillance des releases Helm commençant par 'nac-'")
	for {
		keys := getters.GetAllConfigMapKeys(config.Namespace)
		releases, _ := utils.GetHelmReleases(config.Namespace)
		var nacReleases []string
		for _, r := range releases {
			if len(r) >= 4 && r[:4] == "nac-" {
				nacReleases = append(nacReleases, r[4:]) // On enlève le préfixe 'nac-'
			}
		}
		for _, value := range GetHelmWithoutCM(keys, nacReleases) {
			utils.DeleteHelmRelease(value, config.Namespace)
		}
		time.Sleep(interval)
	}
}
