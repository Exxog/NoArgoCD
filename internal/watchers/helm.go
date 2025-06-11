package watchers

import (
	"time"

	"github.com/Exxog/NoArgoCD/internal/getters"
	"github.com/Exxog/NoArgoCD/internal/utils"
)

// HelmWatcher surveille une release Helm à intervalle régulier
type HelmWatcher struct {
	controller string
	namespace  string
	releaseName string
}

// NewHelmWatcher crée une nouvelle instance de HelmWatcher
func NewHelmWatcher() *HelmWatcher {
	return &HelmWatcher{
		releaseName: "",
		namespace:   "",
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

// Watch lance une boucle infinie qui exécute `helm status` toutes les 30 secondes
func (w *HelmWatcher) WatchOrphelanHelmReleases() {
	for {
		keys := getters.GetAllConfigMapKeys("")
		helm, _ := utils.GetHelmReleasesFiltered("", "nac", "true")
		for _, value := range GetHelmWithoutCM(keys, helm) {
			utils.DeleteHelmRelease(value, "")
		}

		time.Sleep(30 * time.Second)
	}
}
