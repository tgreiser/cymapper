package main

import "github.com/tgreiser/cymapper/cmd/scenebuild/app"

// To use profiler, as described here: https://golang.org/pkg/net/http/pprof/
// uncomment the imports below and the func() in the main function
// import _ "net/http/pprof"
// import "net/http"
// import "log"

func main() {
	// go func() {
		// log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	a := app.Create()
	defer a.RunFinalizers()
	if a != nil {
		err := a.Run()
		if err != nil {
			panic(err)
		}
	}
}
