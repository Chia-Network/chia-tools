package main

import (
	"github.com/chia-network/chia-tools/cmd"
	_ "github.com/chia-network/chia-tools/cmd/certs"
	_ "github.com/chia-network/chia-tools/cmd/config"
	_ "github.com/chia-network/chia-tools/cmd/datalayer"
)

func main() {
	cmd.Execute()
}
