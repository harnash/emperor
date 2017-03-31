package main

import "github.com/harnash/emperor/cmd"

var (
	Version   string
	BuildTime string
)

func main() {
	cmd.Execute()
}
