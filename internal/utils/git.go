package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CloneBranchRepo(repoURL, destination, branch, username, token string) error {
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	}

	if username != "" && token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}

	_, err := git.PlainClone(destination, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors du clonage du dépôt: %v", err)
	}

	fmt.Println("✅ Dépôt cloné avec succès dans", destination)
	return nil
}

func CloneRepo(repoURL, destination, username, token string) error {
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	if username != "" && token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}

	_, err := git.PlainClone(destination, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors du clonage du dépôt: %v", err)
	}

	fmt.Println("✅ Dépôt cloné avec succès dans", destination)
	return nil
}

func CloneOrUpdateRepo(repoURL, destination, branch, username, token string) error {
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		// Si le dossier existe, ouvrir le dépôt et faire un pull
		repo, err := git.PlainOpen(destination)
		if err != nil {
			return fmt.Errorf("❌ Erreur lors de l'ouverture du dépôt: %v", err)
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("❌ Erreur lors de la récupération du worktree: %v", err)
		}

		pullOptions := &git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(branch),
		}

		if username != "" && token != "" {
			pullOptions.Auth = &http.BasicAuth{
				Username: username,
				Password: token,
			}
		}

		err = worktree.Pull(pullOptions)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("❌ Erreur lors du pull du dépôt: %v", err)
		}

		fmt.Println("✅ Dépôt mis à jour avec succès dans", destination)
		return nil
	}

	// Si le dossier n'existe pas, cloner le dépôt
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	}

	if username != "" && token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}

	_, err := git.PlainClone(destination, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors du clonage du dépôt: %v", err)
	}

	fmt.Println("✅ Dépôt cloné avec succès dans", destination)
	return nil
}

func CleanFolderName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	cleaned := re.ReplaceAllString(name, "-")
	cleaned = strings.Trim(cleaned, "-")
	return filepath.Clean(cleaned)
}
