package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func WriteYAMLToFile(filename string, yamlData []byte) error {
	// Convertir en YAML
	///yamlData, err := yaml.Marshal(data)
	//if err != nil {
	//	return fmt.Errorf("erreur de conversion en YAML : %w", err)
	//}

	// Créer ou ouvrir le fichier
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur de création du fichier : %w", err)
	}
	defer file.Close() // Ferme le fichier à la fin

	// Écrire les données YAML dans le fichier
	_, err = file.Write(yamlData)
	if err != nil {
		return fmt.Errorf("erreur d'écriture dans le fichier : %w", err)
	}

	return nil // Pas d'erreur
}

func ConvertToYaml(data map[interface{}]interface{}) []byte {
	yamlData := []byte{}

	if values, exists := data["values"]; exists {
		if data, err := yaml.Marshal(values); err == nil {
			yamlData = data
		} else {
			fmt.Println("Erreur de conversion en YAML :", err)
			return []byte{}
		}

	}
	return yamlData
}
