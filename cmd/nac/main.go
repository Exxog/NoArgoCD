package main

import (
	"log"
	"time"

	"github.com/Exxog/NoArgoCD/internal/controllers"
)

func main() {
	// Création du contrôleur GitLab

	controllerGit := controllers.NewControllerGit(nil)
	controllerHelm := controllers.NewControllerHelm(controllerGit)
	controllerGit.SetHelmController(controllerHelm)

	// Ajout des dépôts à surveiller dans GitLab

	// Démarrer la surveillance des dépôts GitLab
	go controllerGit.StartWatching(30 * time.Second)

	// Création du contrôleur Kube avec une instance de ControllerGit
	controllerKube, err := controllers.NewControllerKube(controllerHelm)
	if err != nil {
		log.Fatalf("❌ Erreur lors de la création du contrôleur Kube : %v", err)
	}

	// Démarrer la surveillance des ConfigMaps dans Kubernetes (dans le namespace "default")
	go controllerKube.StartWatching("")

	// Garder l'application active pour tester
	select {}
}
