package subexcluded

import (
	"github.com/amineck/go-arch-lint/test/check/project/internal/a"
	"github.com/amineck/go-arch-lint/test/check/project/internal/b"
)

func E1() {
	a.A1() // not allowed, but not checked by excluded dir
	b.B1() // not allowed, but not checked by excluded dir
}
