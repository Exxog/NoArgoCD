package watchers

import (
	"fmt"
	"log"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GitLabRepo représente un dépôt GitLab à surveiller
type GitLabRepo struct {
	URL    string
	Branch string
}

// Watcher interface pour gérer la détection de nouveaux commits
type Watcher interface {
	NotifyNewCommit(repo GitLabRepo, commitID string)
}

// GitLabWatcher surveille les commits sur des dépôts GitLab
type GitLabWatcher struct {
	controller   Watcher
	client       *gitlab.Client
	repositories []GitLabRepo
}

// NewGitLabWatcher crée un watcher attaché à un contrôleur et un client GitLab
func NewGitLabWatcher(controller Watcher, client *gitlab.Client) *GitLabWatcher {
	return &GitLabWatcher{
		controller: controller,
		client:     client,
	}
}

// AddRepository permet d'ajouter un dépôt à surveiller
func (w *GitLabWatcher) AddRepository(repo GitLabRepo) {
	w.repositories = append(w.repositories, repo)
	fmt.Printf("📌 Dépôt ajouté : %s (%s)\n", repo.URL, repo.Branch)
	fmt.Println(w.repositories)
}

func checkrepo() {

}

// Watch lance la surveillance des dépôts GitLab
func (w *GitLabWatcher) Watch(interval time.Duration) {
	commitHistory := make(map[string]string)

	for {
		for _, repo := range w.repositories {
			// Récupérer l'ID du projet et la visibilité à partir de l'URL
			projectID, err := getGitLabProjectID(w.client, repo.URL)
			if err != nil {
				log.Printf("❌ Erreur récupération projet %s : %v\n", repo.URL, err)
				continue
			}

			// Récupérer les commits du dépôt
			commits, _, err := w.client.Commits.ListCommits(projectID, &gitlab.ListCommitsOptions{
				RefName: &repo.Branch,
			})
			if err != nil {
				log.Printf("❌ Erreur commits %s [%s] : %v\n", repo.URL, repo.Branch, err)
				continue
			}

			// Vérifier s'il y a un nouveau commit
			if len(commits) > 0 {
				latestCommit := commits[0].ID
				if commitHistory[repo.URL] != latestCommit {
					w.controller.NotifyNewCommit(repo, latestCommit)
					commitHistory[repo.URL] = latestCommit
				}
			} else {
				log.Printf("⚠️ Aucun commit trouvé pour %s [%s]\n", repo.URL, repo.Branch)
			}
		}
		time.Sleep(interval)
	}
}

// getGitLabProjectID récupère l'ID d'un projet GitLab depuis son URL
func getGitLabProjectID(client *gitlab.Client, url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("URL GitLab invalide : %s", url)
	}
	projectName := strings.Join(parts[len(parts)-2:], "/")
	project, _, err := client.Projects.GetProject(projectName, nil)
	if err != nil {
		return "", fmt.Errorf("projet non trouvé : %v", err)
	}
	return fmt.Sprintf("%d", project.ID), nil
}
