package main

import "github.com/tgreiser/cymapper/cmd/scenebuild/app"

func main() {
	a := app.Create()
	defer a.RunFinalizers()
	if a != nil {
		err := a.Run()
		if err != nil {
			panic(err)
		}
	}
}
