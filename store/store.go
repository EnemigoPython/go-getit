package store

import (
	"fmt"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

func OpenStore() {
	filename := runtime.Config.StoreName
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fmt.Println("File opened or created successfully:", filename)
}
