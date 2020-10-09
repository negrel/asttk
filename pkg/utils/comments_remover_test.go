package utils

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/negrel/asttk/pkg/inspector"
	"github.com/negrel/asttk/pkg/parse"
)

func TestRemoveComments(t *testing.T) {
	pkg, err := parse.Package(
		filepath.Join("_data", "comments", "remover"),
		false,
	)
	assert.Nil(t, err, err)

	expectedPkg, err := parse.Package(
		filepath.Join("_data", "comments", "remover_expected"),
		false,
	)
	assert.Nil(t, err)

	editor := inspector.New(
		RemoveComments(func(_ string) bool { return true }),
	)
	for i, file := range pkg.Files {
		editor.Inspect(file.AST())

		expectedResult, err := expectedPkg.Files[i].Bytes()
		assert.Nil(t, err, err)

		actualResult, err := file.Bytes()
		assert.Nil(t, err, err)
		fmt.Println(string(actualResult))

		assert.EqualValues(t, string(expectedResult), string(actualResult))
	}
}
