// Invokes the perigord driver application

package main

import (
	_ "{{.project}}/migrations"
	"gitlab.inn4science.com/gophers/perigord/stub"
)

func main() {
	stub.StubMain()
}
