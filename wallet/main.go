package main

import "github.com/Crows-Storm/Axis/common/config"

func init() {
	if err := config.NewViperConfig(); err != nil {
		panic("Init ViperConfig ERROR !!!")
	}
}

func main() {

}
