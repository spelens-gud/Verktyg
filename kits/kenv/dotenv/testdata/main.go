package main

import (
	"fmt"
	"os"

	"git.bestfulfill.tech/devops/go-core/kits/kenv/dotenv"

	_ "git.bestfulfill.tech/devops/go-core/kits/kenv/dotenv/testdata/test"
)

var _ = dotenv.ImportMe

func main() {
	fmt.Printf(os.Getenv("TEST_ENV"))
}
