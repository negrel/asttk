package edit

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/negrel/asttk/pkg/inspector"
	"github.com/negrel/asttk/pkg/parse"
)

type unusedImportsRemoverTest struct {
	src string
	out string
}

var unusedImportsRemoverTests = []unusedImportsRemoverTest{
	// variable identifier that shadow a package name
	{
		src: `package main

import (
	"fmt"
	"log"
	// image is imported but not used
	"image"
)

type date struct {
	dd, mm, yy string 
}

func (d date) String() {
	return string(d.dd)
}

func main() {
	image := date{
		dd:	"01",
		mm: "01",
		yy:	"1970",
	}

	log.Println("Hello world")
	greet(image)
}

func greet(a fmt.Stringer) {
	log.Println("Hello", a)
}

`,
		out: `package main

import (
	"fmt"
	"log"
)

type date struct { 
	dd, mm, yy string
}

func (d date) String() {
	return string(d.dd)
}


func main() {
	image := date{
		dd:	"01",
		mm: "01",
		yy:	"1970",
	}

	log.Println("Hello world")
	greet(image)
}

func greet(a fmt.Stringer) {
	log.Println("Hello", a)
}
`},
	// fff is an identifier, so the remover must avoid ast.GenDecl with an IMPORT token.
	{
		src: `package main

import (
	fff "fmt"
)

func main() {
	fff.Println("Hello world")
}
`,
		out: `package main

import (
	fff "fmt"
)

func main() {
	fff.Println("Hello world")
}
`},
	// fmt identifier is used twice but the package is imported once.
	{
		src: `package main

import (
	"fmt"
)

func main() {
	fmt.Print("Hello")
	fmt.Println(" world")
}
`,
		out: `package main

import (
	"fmt"
)

func main() {
	fmt.Print("Hello")
	fmt.Println(" world")
}
`},
}

func TestUnusedImportsRemover(t *testing.T) {
	fset := token.NewFileSet()

	findUnusedImports, removeUnusedImports := RemoveUnusedImports()
	editor := inspector.New(findUnusedImports)

	for _, test := range unusedImportsRemoverTests {
		file, err := parser.ParseFile(fset, "unusedImportsRemover", test.src, parser.AllErrors)
		assert.Nil(t, err)

		editor.Inspect(file)
		removeUnusedImports(file)

		actualResult, err := parse.NewGoFile("", file).Byte()
		assert.Nil(t, err)

		expectedFile, err := parser.ParseFile(fset, "unusedImportsRemover", test.out, parser.AllErrors)
		assert.Nil(t, err)

		expectedResult, err := parse.NewGoFile("", expectedFile).Byte()
		assert.Nil(t, err)

		assert.EqualValues(t, string(expectedResult), string(actualResult))
	}
}
