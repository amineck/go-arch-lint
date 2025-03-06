package assembler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/amineck/go-arch-lint/internal/models"
)

type resolver struct {
	pathResolver  pathResolver
	rootDirectory string
	moduleName    string
}

func newResolver(
	pathResolver pathResolver,
	rootDirectory string,
	moduleName string,
) *resolver {
	return &resolver{
		pathResolver:  pathResolver,
		rootDirectory: rootDirectory,
		moduleName:    moduleName,
	}
}

func (r *resolver) resolveLocalGlobPath(localGlobPath string) ([]models.ResolvedPath, error) {
	list := make([]models.ResolvedPath, 0)

	absPath := fmt.Sprintf("%s/%s", r.rootDirectory, localGlobPath)
	resolved, err := r.pathResolver.Resolve(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path '%s'", absPath)
	}

	for _, absResolvedPath := range resolved {
		localPath := strings.TrimPrefix(absResolvedPath, fmt.Sprintf("%s/", r.rootDirectory))
		localPath = strings.TrimRight(localPath, "/")
		importPath := fmt.Sprintf("%s/%s", r.moduleName, localPath)

		list = append(list, models.ResolvedPath{
			ImportPath: strings.TrimRight(importPath, "/"),
			LocalPath:  strings.TrimRight(localPath, "/"),
			AbsPath:    filepath.Clean(strings.TrimRight(absResolvedPath, "/")),
		})
	}

	return list, nil
}
