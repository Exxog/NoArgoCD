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

	// Création du contrôleur
	controller := controllers.NewController(client)

	// Ajout des dépôts à surveiller
	controller.AddRepository("https://gitlab.com/gitlab-org/gitlab", "master")
	controller.AddRepository("https://gitlab.com/gitlab-org/gitlab-runner", "main")

	// Démarrer la surveillance
	interval := 30 * time.Second
	controller.StartWatching(interval)
}
