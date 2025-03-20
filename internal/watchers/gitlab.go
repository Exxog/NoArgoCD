package watchers

import (
	"fmt"
	"log"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Configuration du dépôt GitLab avec URL
type GitLabRepo struct {
	URL    string
	Branch string
}

// Fonction pour récupérer l'ID et la visibilité d'un projet GitLab à partir de son URL
func getGitLabProjectDetails(client *gitlab.Client, url string) (string, gitlab.VisibilityValue, error) {
	// Extraire le nom du projet à partir de l'URL
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("URL GitLab invalide : %s", url)
	}

	// Le projet doit être au format `namespace/project_name`, donc concaténons le namespace et le nom du projet
	projectName := strings.Join(parts[len(parts)-2:], "/")

	// Récupérer les détails du projet à partir du nom complet du projet
	project, _, err := client.Projects.GetProject(projectName, nil)
	if err != nil {
		return "", "", fmt.Errorf("❌ Erreur lors de la récupération du projet : %v", err)
	}

	// Retourner l'ID du projet et sa visibilité
	return fmt.Sprintf("%d", project.ID), project.Visibility, nil
}

// Fonction pour créer un client GitLab, utilisant un token si nécessaire
func createGitLabClient(token string) (*gitlab.Client, error) {
	// Si un token est fourni, l'utiliser, sinon créer un client sans token
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client GitLab : %v", err)
	}
	return client, nil
}

// Interface Watcher pour notifier des nouveaux commits
type Watcher interface {
	NotifyNewCommit(repo GitLabRepo, commitID string)
}

// Controller qui reçoit les notifications de commit
type Controller struct{}

// Méthode qui gère les notifications de nouveaux commits
func (c *Controller) NotifyNewCommit(repo GitLabRepo, commitID string) {
	fmt.Printf("📝 Nouveau commit détecté dans %s (%s) : %s\n", repo.URL, repo.Branch, commitID)
}

// Watcher GitLab pour surveiller les commits
type GitLabWatcher struct {
	controller Watcher
}

// Nouveau GitLabWatcher avec le controller attaché
func NewGitLabWatcher(controller Watcher) *GitLabWatcher {
	return &GitLabWatcher{controller: controller}
}

// Fonction pour surveiller les commits des dépôts GitLab
func (watcher *GitLabWatcher) WatchGitLabRepos(client *gitlab.Client, repos []GitLabRepo, interval time.Duration) {
	commitHistory := make(map[string]string)

	for {
		for _, repo := range repos {
			// Récupérer l'ID du projet et la visibilité à partir de l'URL
			projectID, visibility, err := getGitLabProjectDetails(client, repo.URL)
			if err != nil {
				log.Printf("❌ Erreur lors de la récupération des informations du projet pour %s : %v\n", repo.URL, err)
				continue
			}

			// Si le projet est privé, créer un nouveau client avec le token
			if visibility == gitlab.PrivateVisibility {
				fmt.Printf("⚠️ Le projet %s est privé. Utilisation du token...\n", repo.URL)
				token := "ton-token-ici" // Remplacer par ton token GitLab
				client, err = createGitLabClient(token)
				if err != nil {
					log.Printf("❌ Erreur lors de la création du client avec le token : %v\n", err)
					continue
				}
			}

			// Débogage : Vérification du projectID avant de récupérer les commits
			fmt.Printf("🔍 Récupération des commits pour le projet ID %s, branche %s...\n", projectID, repo.Branch)

			// Récupérer les commits du dépôt
			commits, _, err := client.Commits.ListCommits(projectID, &gitlab.ListCommitsOptions{
				RefName: &repo.Branch,
			})
			if err != nil {
				log.Printf("❌ Erreur sur %s [%s] : %v\n", repo.URL, repo.Branch, err)
				continue
			}

			// Si aucun commit n'est trouvé, afficher un message
			if len(commits) == 0 {
				log.Printf("❌ Aucun commit trouvé pour %s [%s].\n", repo.URL, repo.Branch)
				continue
			}

			latestCommit := commits[0].ID
			if commitHistory[repo.URL] != latestCommit {
				// Notifier le contrôleur si un nouveau commit est trouvé
				watcher.controller.NotifyNewCommit(repo, latestCommit)
				commitHistory[repo.URL] = latestCommit
			}
		}
		time.Sleep(interval)
	}
}

func main() {
	// Création du client GitLab sans authentification (accès public pour les projets publics)
	client, err := createGitLabClient("")
	if err != nil {
		log.Fatalf("❌ Erreur lors de la création du client GitLab: %v", err)
	}

	// Liste des URLs des dépôts GitLab à surveiller
	repos := []GitLabRepo{
		{URL: "https://gitlab.com/gitlab-org/gitlab", Branch: "master"},
		{URL: "https://gitlab.com/gitlab-org/gitlab-runner", Branch: "main"},
	}

	// Création du contrôleur pour recevoir les notifications de commit
	controller := &Controller{}
	// Création du watcher GitLab
	watcher := NewGitLabWatcher(controller)

	// Intervalle pour vérifier les commits
	interval := 30 * time.Second
	fmt.Println("🔍 Démarrage du watcher GitLab...")
	// Lancer la surveillance des commits
	watcher.WatchGitLabRepos(client, repos, interval)
}
