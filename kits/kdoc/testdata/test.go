package testdata

import "git.bestfulfill.tech/devops/go-core/kits/kdoc/testdata/test2"

type Test struct {
	AA string // 注释AA
	ETest
	test2.Test
}

type ETest struct {
	ETest string // E嵌套
}
