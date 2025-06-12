package main

import (
	"log"
	"time"

	"github.com/Exxog/NoArgoCD/internal/config"
	"github.com/Exxog/NoArgoCD/internal/controllers"
	"github.com/Exxog/NoArgoCD/internal/utils"
)

func main() {
	config.Namespace = utils.GetNamespace(config.Namespace)
	// Création du contrôleur GitLab

	controllerGit := controllers.NewControllerGit(nil)
	controllerHelm := controllers.NewControllerHelm(controllerGit)
	controllerGit.SetHelmController(controllerHelm)

	// Démarrer la surveillance des dépôts Git
	go controllerGit.StartWatching(30 * time.Second)

	// Création du contrôleur Kube avec une instance de ControllerGit
	controllerKube, err := controllers.NewControllerKube(controllerHelm)
	if err != nil {
		log.Fatalf("❌ Erreur lors de la création du contrôleur Kube : %v", err)
	}

	// Démarrer la surveillance des ConfigMaps dans Kubernetes (dans le namespace "default")
	go controllerKube.StartWatching(config.Namespace)

	// Démarrer la suppressions des release helms orphelines
	go controllerHelm.StartWatching(30 * time.Second)

	// Garder l'application active pour tester
	select {}
}
