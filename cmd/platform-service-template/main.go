//go:generate ocp-gen
package main

import (
	"fmt"
	"os"

	// ocp-gen:replace github.com/openmcp-project/platform-service-template=MODULE platform-service-template=CMD_FOLDER
	"github.com/openmcp-project/platform-service-template/cmd/platform-service-template/app"
)

func main() {
	cmd := app.NewPlatformServiceCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
