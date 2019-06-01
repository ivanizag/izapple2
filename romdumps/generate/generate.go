// To generate the resources put the files on a "files" subdirectory and run main

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()
	templates := http.Dir(filepath.Join(cwd, "files"))
	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     "../romdumps_vfsdata.go",
		PackageName:  "romdumps",
		VariableName: "Assets",
	}); err != nil {
		log.Fatalln(err)
	}
}
