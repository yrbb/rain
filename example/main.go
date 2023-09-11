package main

import (
	"github.com/yrbb/rain"

	_ "github.com/yrbb/rain/example/cmd"
)

func main() {
	app, err := rain.New()
	if err != nil {
		panic(err)
	}

	app.OnStart(func() {
		// do something
		// fmt.Println("on start...")
	})

	app.OnStop(func() {
		// do something
		// fmt.Println("on stop...")
	})

	app.OnConfigUpdate(func(config *rain.Config) {
		// if reflect.DeepEqual(config.Custom, app.Config.Custom) {
		// 	return
		// }

		// for k, v := range config.Custom {
		// 	fmt.Printf("config update, key: %s, value: %v\n", k, v)
		// }
	})

	app.Run()
}
