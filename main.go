package main

import (
	_ "embed"
	"github.com/sdesignb/polygon-edge/command/root"
	"github.com/sdesignb/polygon-edge/licenses"
)

var (
	//go:embed LICENSE
	license string
)

func main() {
	licenses.SetLicense(license)

	root.NewRootCommand().Execute()
}
