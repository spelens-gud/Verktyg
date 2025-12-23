# 环境变量加载器

从`.env`文件加载环境变量 应放置于二进制入口import 且放置于更靠前(影响import加载顺序)

```go
package main

import (
	"github.com/spelens-gud/Verktyg/kits/kenv/dotenv"

	"fmt"
	"os"
    
    "others/imports"
)

var _ = dotenv.ImportMe

func main() {
	fmt.Printf(os.Getenv("TEST_ENV"))
}
```