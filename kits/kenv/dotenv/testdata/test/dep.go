package test

import (
	"fmt"
	"os"
)

func init() {
	fmt.Println(os.Getenv("TEST_ENV"), "load")
}
