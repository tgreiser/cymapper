package main

import "github.com/tgreiser/cymapper/cmd/scenebuild/app"

func main() {
	a := app.Create()
	if a != nil {
		err := a.Run()
		if err != nil {
			panic(err)
		}
	}
}
