package main

import (
	"github.com/harnash/emperor/cmd"
	"github.com/spf13/viper"
)

var (
	Version   string
	BuildTime string
)

func main() {
	viper.Set("version", Version)
	viper.Set("build_time", BuildTime)

	cmd.Execute()
}
