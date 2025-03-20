package watchers

import (
	"context"
	"fmt"
	"log"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/configmap"

	"github.com/Exxog/NoArgoCD/internal/controllers"
)

// KubeWatcher surveille les ConfigMaps pour récupérer les dépôts GitLab
type KubeWatcher struct {
	client     *kubernetes.Clientset
	controller *controllers.ControllerGit
}

// NewKubeWatcher crée un watcher pour écouter les ConfigMaps
func NewKubeWatcher(controller *controllers.ControllerGit) (*KubeWatcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur de configuration du client Kubernetes: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("❌ Erreur de création du client Kubernetes: %v", err)
	}

	return &KubeWatcher{
		client:     clientset,
		controller: controller,
	}, nil
}

// Watch écoute les ConfigMaps et informe le ControllerGit
func (w *KubeWatcher) Watch(namespace string) {
	ctx := context.Background()
	configMapWatcher := configmap.NewInformedWatcher(w.client, namespace)

	configMapWatcher.Watch("gitlab-repos", func(cm *v1.ConfigMap) {
		if cm == nil {
			log.Println("⚠️ ConfigMap supprimée ! Arrêt de la surveillance.")
			return
		}

		fmt.Println("🔄 Mise à jour des dépôts GitLab à partir de ConfigMap !")

		var repos []controllers.GitLabRepo

		// Extraction des dépôts
		for key, value := range cm.Data {
			parts := strings.Split(value, ",")
			if len(parts) != 2 {
				log.Printf("⚠️ Format invalide pour %s : %s\n", key, value)
				continue
			}

			repos = append(repos, controllers.GitLabRepo{
				URL:    parts[0],
				Branch: parts[1],
			})
		}

		// Envoyer les nouvelles données au ControllerGit
		w.controller.UpdateRepos(repos)
	})

	// Démarre l'écoute
	if err := configMapWatcher.Start(ctx.Done()); err != nil {
		log.Fatalf("❌ Erreur lors de l'écoute des ConfigMaps: %v", err)
	}
}
