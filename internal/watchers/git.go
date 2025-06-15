package watchers

import (
	"fmt"
	"log"
	"time"

	"github.com/Exxog/NoArgoCD/internal/config"
	"github.com/Exxog/NoArgoCD/internal/utils"
)

// gitRepo représente un dépôt git à surveiller
type GitRepo struct {
	URL            string
	Branch         string
	AuthSecretName string
}

// Watcher interface pour gérer la détection de nouveaux commits
type GitController interface {
	NotifyNewCommit(repo GitRepo, commitID string)
}

// gitWatcher surveille les commits sur des dépôts git
type GitWatcher struct {
	gitController   GitController
	repositories    []GitRepo
	repositoriesMap map[string]struct{}
}

// NewgitWatcher crée un watcher attaché à un contrôleur et un client git
func NewGitWatcher(gitController GitController) *GitWatcher {
	return &GitWatcher{
		gitController:   gitController,
		repositoriesMap: make(map[string]struct{}),
	}
}

func (w *GitWatcher) AddRepository(repo GitRepo) {
	// Créer une clé unique pour chaque dépôt basée sur l'URL et la branche
	key := fmt.Sprintf("%s-%s", repo.URL, repo.Branch)

	// Vérifier si le dépôt existe déjà en utilisant la map
	if _, exists := w.repositoriesMap[key]; exists {
		// Si le dépôt existe déjà, afficher un message
		fmt.Printf("[watchers][git] ⚠️ Le dépôt %s (%s) existe déjà !\n", repo.URL, repo.Branch)
		return
	}

	// Si le dépôt n'existe pas, l'ajouter à la liste et à la map
	w.repositories = append(w.repositories, repo)
	w.repositoriesMap[key] = struct{}{}
	fmt.Printf("[watchers][git] 📌 Dépôt ajouté : %s (%s)\n", repo.URL, repo.Branch)
}

func (w *GitWatcher) RemoveRepository(repo GitRepo) {
	// Créer une clé unique pour identifier le dépôt
	key := fmt.Sprintf("%s-%s", repo.URL, repo.Branch)

	// Vérifier si le dépôt existe dans la map
	if _, exists := w.repositoriesMap[key]; !exists {
		fmt.Printf("[watchers][git] ⚠️ Le dépôt %s (%s) n'existe pas dans la liste !\n", repo.URL, repo.Branch)
		return
	}

	// Supprimer de la map
	delete(w.repositoriesMap, key)

	// Supprimer de la slice w.repositories
	for i, r := range w.repositories {
		if r.URL == repo.URL && r.Branch == repo.Branch {
			// Retirer l'élément de la slice
			w.repositories = append(w.repositories[:i], w.repositories[i+1:]...)
			fmt.Printf("[watchers][git] ❌ Dépôt supprimé : %s (%s)\n", repo.URL, repo.Branch)
			return
		}
	}
}

func (w *GitWatcher) NotifyNewCommit(repo GitRepo, commit string) {
	// Implémente la logique pour notifier d'un nouveau commit, par exemple un log ou un appel à une API
	log.Printf("[watchers][git] Nouveau commit détecté dans le dépôt %s (branche: %s): %s\n", repo.URL, repo.Branch, commit)
	w.gitController.NotifyNewCommit(repo, commit)

}

func (w *GitWatcher) CheckRepo(repo GitRepo, commitHistory map[string]string, secretsNames map[string][]string) {
	// Récupérer le dernier commit du dépôt distant
	user, token, _ := utils.GetUsernamePasswordFromSecret(config.Namespace, repo.AuthSecretName)
	latestCommit, err := utils.GetLatestCommit(repo.URL, repo.Branch, user, token)
	if err != nil {
		log.Printf("[watchers][git] ❌ Erreur lors de la récupération du dernier commit %s [%s]: %v\n", repo.URL, repo.Branch, err)
		return
	}

	// Vérifier si le commit a changé
	if commitHistory[repo.URL+"-"+repo.Branch] != latestCommit {
		// Si un nouveau commit est trouvé, notifie
		w.NotifyNewCommit(repo, latestCommit)
		// Met à jour l'historique des commits
		commitHistory[repo.URL+"-"+repo.Branch] = latestCommit
	} else {
		log.Println("[watchers][git]", repo.URL, repo.Branch, " last:", latestCommit, " current:", commitHistory[repo.URL+"-"+repo.Branch], " Pas de nouveau commit.")
	}

}

func (w *GitWatcher) WatchRepo(interval time.Duration) {
	commitHistory := make(map[string]string)
	secretsNames := make(map[string][]string)

	// Lancer la surveillance des dépôts à intervalles réguliers
	for {
		for _, repo := range w.repositories {
			w.CheckRepo(repo, commitHistory, secretsNames)
		}
		time.Sleep(interval)
	}
}
