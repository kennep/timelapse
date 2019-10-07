// +build ignore

package main

import (
	"log"

	"github.com/kennep/timelapse/endpoints"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(endpoints.Assets, vfsgen.Options{
		PackageName:  "endpoints",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
