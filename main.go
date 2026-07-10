package main

import "github.com/KevinLanahan/lokal/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
