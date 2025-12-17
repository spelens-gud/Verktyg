package testdata

import "github.com/spelens-gud/Verktyg.git/kits/kdoc/testdata/test2"

type Test struct {
	AA string // 注释AA
	ETest
	test2.Test
}

type ETest struct {
	ETest string // E嵌套
}
