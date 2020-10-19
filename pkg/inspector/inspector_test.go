package inspector

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
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
			log.Println("decl:", c.value)

			return false
		}

		return true
	}
}

func stmtCount(c *counter) Inspector {
	return func(node ast.Node) bool {
		if _, isStmt := node.(ast.Stmt); isStmt {
			c.value++
			log.Println("stmt:", c.value, node)
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
	log.Println("expected")
	ast.Inspect(file, declCount(expectedDeclCounter))
	ast.Inspect(file, nothingCount(expectedNothingCounter))

	// Actual
	declCounter := new(counter)
	nothingCounter := new(counter)
	log.Println("actual")
	l1 := Lieutenant(declCount(declCounter), nothingCount(nothingCounter))
	l2 := Lieutenant(declCount(declCounter), nothingCount(nothingCounter))
	lInspector := New(
		l1,
		l2,
	)

	log.Println("l1", l1)
	log.Println("l2", l2)

	lInspector.Inspect(file)

	assert.Equal(t, expectedDeclCounter.value*2, declCounter.value)
	assert.Equal(t, expectedNothingCounter.value*2, nothingCounter.value)
}

func disableAfterNthNode(nth int, recorder *[]int) Inspector {
	c := nth

	return func(node ast.Node) bool {
		c--
		if c == 0 {
			c = nth
			return false
		}

		*recorder = append(*recorder, nth)
		return true
	}
}

func TestLead_OrderRemain(t *testing.T) {
	testAST := &ast.BlockStmt{
		List: []ast.Stmt{},
	}

	for i := 0; i < 127; i++ {
		testAST.List = append(testAST.List, &ast.ExprStmt{
			X: ast.NewIdent(fmt.Sprint(i)),
		})
	}

	var recorder []int
	inspectors := make([]Inspector, 8)
	for i := 0; i < 8; i++ {
		inspectors[i] = disableAfterNthNode(i+1, &recorder)
	}

	inspectors = append(inspectors, func(node ast.Node) bool {
		recorder = append(recorder, -1)

		return true
	})

	lInspector := New(inspectors...)
	lInspector.Inspect(testAST)

	previousRecord := recorder[0]
	for index, record := range recorder[1:] {
		if record != -1 && record <= previousRecord {
			t.Logf("Lead inspector swap recorder (inspector) %v and %v. (index %v)", record, previousRecord, index)
			t.Fail()
		}
		previousRecord = record
	}
}
