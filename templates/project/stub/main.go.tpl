// Invokes the perigord driver application

package main

import (
	_ "{{.project}}/migrations"
	"github.com/inn4science/perigord/stub"
)

func main() {
	stub.StubMain()
}
