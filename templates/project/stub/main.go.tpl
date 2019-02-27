// Invokes the perigord driver application

package main

import (
	_ "{{.project}}/migrations"
	"gitlab.com/go-truffle/enhanced-perigord/stub"
)

func main() {
	stub.StubMain()
}
