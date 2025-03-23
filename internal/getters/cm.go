package getters

import (
	"fmt"
	"log"

	"github.com/Exxog/NoArgoCD/internal/utils"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

func findHelmEntriesWithRepoURL(configMaps []v1.ConfigMap, targetURL, typeGitOPS, targetRevision string) map[string][]map[interface{}]interface{} {
	result := make(map[string][]map[interface{}]interface{})

	for _, cm := range configMaps {
		for key, data := range cm.Data {
			var parsedData map[string]interface{}
			if err := yaml.Unmarshal([]byte(data), &parsedData); err != nil {
				fmt.Printf("❌ Erreur de parsing YAML pour ConfigMap '%s' (clé: '%s'): %v\n", cm.Name, key, err)
				continue
			}

			helmData, ok := parsedData[typeGitOPS].(map[interface{}]interface{})
			if !ok {
				fmt.Println("error")
				continue
			}

			repoURL, repoFound := helmData["repoUrl"].(string)
			revision, revFound := helmData["targetRevision"].(string)

			if repoFound && revFound && repoURL == targetURL && revision == targetRevision {
				result[key] = append(result[key], helmData)
				fmt.Printf("✅ ConfigMap '%s' (clé: '%s') contient le repoUrl et targetRevision attendus.\n", cm.Name, key)
			}
		}
	}
	return result
}
func getConfigMaps(namespace, label string) []v1.ConfigMap {
	clientset, err := utils.SetupKubernetesClient()
	if err != nil {
		log.Fatalf("Erreur lors de la configuration du client Kubernetes: %v", err)
	}
	namespace = utils.GetNamespace(namespace)

	configMaps, err := utils.GetConfigMapsWithLabel(clientset, namespace, label)
	if err != nil {
		log.Fatalf("Erreur lors de la récupération des ConfigMaps: %v", err)
	}

	return configMaps
}

func GetHelm(targetURL, targetRevision, namespace string) map[string][]map[interface{}]interface{} {
	configMaps := getConfigMaps(namespace, "nac=true")
	return findHelmEntriesWithRepoURL(configMaps, targetURL, "helm", targetRevision)

}

func GetAllConfigMapKeys(namespace string) []string {

	configMaps := getConfigMaps(namespace, "nac=true")
	var keys []string
	uniqueKeys := make(map[string]struct{}) // Pour éviter les doublons

	for _, cm := range configMaps {
		for key := range cm.Data {
			if _, exists := uniqueKeys[key]; !exists {
				keys = append(keys, key)
				uniqueKeys[key] = struct{}{}
			}
		}
	}

	return keys
}
