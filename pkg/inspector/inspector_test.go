package inspector

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

type counter struct {
	value int
}

var helloWorld = `
	package main

	import "fmt"

	func main() {
		greet("World")
	}

	func greet(name string) {
		fmt.Println("Hello", name)
	}
`

func declCount(c *counter) Inspector {
	return func(node ast.Node) bool {
		if _, isDecl := node.(ast.Decl); isDecl {
			c.value++
			return false
		}

		return true
	}
}

func stmtCount(c *counter) Inspector {
	return func(node ast.Node) bool {
		if _, isStmt := node.(ast.Stmt); isStmt {
			c.value++
		}

		return true
	}
}

func TestLead_InspectAll(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", helloWorld, parser.AllErrors)
	assert.Nil(t, err)

	// Expected
	expectedDeclCounter := new(counter)
	expectedStmtCounter := new(counter)
	ast.Inspect(file, declCount(expectedDeclCounter))
	ast.Inspect(file, stmtCount(expectedStmtCounter))

	// Actual
	declCounter := new(counter)
	stmtCounter := new(counter)
	lInspector := New(declCount(declCounter), stmtCount(stmtCounter))
	lInspector.Inspect(file)

	assert.Equal(t, expectedDeclCounter.value, declCounter.value)
	assert.Equal(t, expectedStmtCounter.value, stmtCounter.value)
}

func nothingCount(c *counter) Inspector {
	return func(node ast.Node) bool {
		if node != nil {
			c.value++
		}

		return false
	}
}

func TestLead_InspectNothing(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", helloWorld, parser.AllErrors)
	assert.Nil(t, err)

	// Expected
	expectedCounter := new(counter)
	expectedCount := nothingCount(expectedCounter)
	ast.Inspect(file, expectedCount)

	// Actual
	counters := make([]*counter, 2)
	counts := make([]Inspector, len(counters))
	for i := 0; i < len(counters); i++ {
		counters[i] = new(counter)
		counts[i] = nothingCount(counters[i])
	}

	lInspector := New(counts...)
	lInspector.Inspect(file)

	for i := 0; i < len(counters); i++ {
		assert.Equal(t, expectedCounter.value, counters[i].value)
	}
}

func TestLead_InspectMixed(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", helloWorld, parser.AllErrors)
	assert.Nil(t, err)

	// Expected
	expectedDeclCounter := new(counter)
	expectedNothingCounter := new(counter)
	ast.Inspect(file, declCount(expectedDeclCounter))
	ast.Inspect(file, nothingCount(expectedNothingCounter))

	// Actual
	declCounter := new(counter)
	nothingCounter := new(counter)
	lInspector := New(declCount(declCounter), nothingCount(nothingCounter))
	lInspector.Inspect(file)

	assert.Equal(t, expectedDeclCounter.value, declCounter.value)
	assert.Equal(t, expectedNothingCounter.value, nothingCounter.value)
}

func TestLieutenant(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", helloWorld, parser.AllErrors)
	assert.Nil(t, err)

	// Expected
	expectedDeclCounter := new(counter)
	expectedNothingCounter := new(counter)
	ast.Inspect(file, declCount(expectedDeclCounter))
	ast.Inspect(file, nothingCount(expectedNothingCounter))

	// Actual
	declCounter := new(counter)
	nothingCounter := new(counter)
	lInspector := New(
		Lieutenant(declCount(declCounter), nothingCount(nothingCounter)),
		Lieutenant(declCount(declCounter), nothingCount(nothingCounter)))

	lInspector.Inspect(file)

	assert.Equal(t, (expectedDeclCounter.value * 2), declCounter.value)
	assert.Equal(t, (expectedNothingCounter.value * 2), nothingCounter.value)
}

func recordEveryNthNode(nth int, recorder *[]int) Inspector {
	i := 0
	return func(node ast.Node) bool {
		i++

		if i%nth == 0 {
			*recorder = append(*recorder, nth)
		}

		return true
	}
}

func TestLead_OrderRemain(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", helloWorld, parser.AllErrors)
	assert.Nil(t, err)

	recorder := []int{}
	lInspector := New(
		recordEveryNthNode(3, &recorder),
		recordEveryNthNode(2, &recorder),
		recordEveryNthNode(1, &recorder),
	)
	lInspector.Inspect(file)

	expected := func() []int {
		result := make([]int, 0, 256)
		i := 0
		for len(result) < 256 {
			i++

			if i%3 == 0 {
				result = append(result, 3)
			}
			if i%2 == 0 {
				result = append(result, 2)
			}
			if i%1 == 0 {
				result = append(result, 1)
			}
		}

		return result[:len(recorder)]
	}()

	assert.EqualValues(t, expected, recorder)
}
