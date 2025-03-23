package watchers

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/Exxog/NoArgoCD/internal/getters"
	"github.com/Exxog/NoArgoCD/internal/utils"
)

// HelmWatcher surveille une release Helm à intervalle régulier
type HelmWatcher struct {
	releaseName string
	namespace   string
}

// NewHelmWatcher crée une nouvelle instance de HelmWatcher
func NewHelmWatcher(releaseName, namespace string) *HelmWatcher {
	return &HelmWatcher{
		releaseName: releaseName,
		namespace:   namespace,
	}
}

// ShowRelease affiche les détails de la release Helm
func (w *HelmWatcher) ShowRelease() {
	cmd := exec.Command("helm", "status", w.releaseName, "--namespace", w.namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[watchers][helm] ❌ Erreur lors de l'exécution de 'helm status %s' dans le namespace '%s': %v\n", w.releaseName, w.namespace, err)
		return
	}
	fmt.Printf("[watchers][helm] 📜 Détails de la release %s (namespace: %s):\n%s\n", w.releaseName, w.namespace, string(output))
}

// Watch lance une boucle infinie qui exécute `helm status` toutes les 30 secondes
func (w *HelmWatcher) Watch() {
	for {
		keys := getters.GetAllConfigMapKeys("")
		helm, _ := utils.GetHelmReleases("")

		w.ShowRelease()
		time.Sleep(30 * time.Second)
	}
}
