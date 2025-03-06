package a

import "github.com/amineck/go-arch-lint/test/check/project/internal/common"

func A1() {
	common.C1() // common - allowed
}
