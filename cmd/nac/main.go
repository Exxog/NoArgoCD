package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Exxog/NoArgoCD/internal/controllers"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// createGitLabClient initialise un client GitLab
func createGitLabClient(token string) (*gitlab.Client, error) {
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("erreur création client GitLab : %v", err)
	}
	return client, nil
}

func main() {
	// Création du client GitLab (token facultatif pour projets publics)
	client, err := createGitLabClient("")
	if err != nil {
		log.Fatalf("❌ Impossible de créer le client GitLab : %v", err)
	}

	// Création du contrôleur GitLab

	controllerGit := controllers.NewControllerGit(client, nil)
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
