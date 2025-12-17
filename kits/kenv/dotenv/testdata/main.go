package main

import (
	"fmt"
	"os"

	"github.com/spelens-gud/Verktyg.git/kits/kenv/dotenv"

	_ "github.com/spelens-gud/Verktyg.git/kits/kenv/dotenv/testdata/test"
)

var _ = dotenv.ImportMe

func main() {
	fmt.Printf(os.Getenv("TEST_ENV"))
}
