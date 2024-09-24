package main

import (
	"github.com/chia-network/chia-tools/cmd"
	_ "github.com/chia-network/chia-tools/cmd/certs"
	_ "github.com/chia-network/chia-tools/cmd/config"
	_ "github.com/chia-network/chia-tools/cmd/datalayer"
	_ "github.com/chia-network/chia-tools/cmd/network"
	_ "github.com/chia-network/chia-tools/cmd/testnet"
)

func main() {
	cmd.Execute()
}
