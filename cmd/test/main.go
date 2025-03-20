package main

import (
	"fmt"

	"github.com/Exxog/NoArgoCD/internal/controllers"
)

func main() {
	// Tester directement le ControllerKube
	fmt.Println("🔄 Lancement du test ControllerKube...")
	controllers.RunControllerKubeTest()
}
