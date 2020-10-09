package inspector

import (
	"go/ast"
	"sort"
)

// Inspector define any function that can be used for AST inspection.
type Inspector func(ast.Node) bool

// Lieutenant define an Inspector that manage his own Inspectors.
func Lieutenant(Inspectors ...Inspector) Inspector {
	return New(Inspectors...).inspect
}

// Lead is the Inspector chief that manage the inspection.
type Lead struct {
	depth    int
	active   map[int]Inspector
	inactive map[int]map[int]Inspector
}

// New return an Inspector Lead.
func New(inspectors ...Inspector) *Lead {
	insp := make(map[int]Inspector, len(inspectors))
	for index, inspector := range inspectors {
		insp[index] = inspector
	}

	return &Lead{
		active:   insp,
		depth:    0,
		inactive: make(map[int]map[int]Inspector),
	}
}

// Inspect start the inspection of the given node.
func (l *Lead) Inspect(node ast.Node) {
	l.enableInactive()
	ast.Inspect(node, l.inspect)
}

func (l *Lead) inspect(node ast.Node) bool {
	if node != nil {
		defer func() { l.depth++ }()
	} else {
		defer func() {
			l.depth--
			l.enableInactive()
		}()
	}

	for index, inspector := range l.inspectors() {
		recursiveHook := inspector(node)

		if !recursiveHook {
			l.disableForSubTree(index)
		}
	}

	if len(l.active) == 0 {
		l.enableInactive()
		return false
	}
	return true
}

func (l *Lead) disableForSubTree(index int) {
	if l.inactive[l.depth] == nil {
		l.inactive[l.depth] = make(map[int]Inspector)
	}

	l.inactive[l.depth][index] = l.active[index]
	delete(l.active, index)
}

func (l *Lead) enableInactive() {
	if _, ok := l.inactive[l.depth]; !ok {
		return
	}

	for index, inspector := range l.inactive[l.depth] {
		l.active[index] = inspector
	}
	delete(l.inactive, l.depth)
}

// return an ordered array of the active inspectors
func (l *Lead) inspectors() []Inspector {
	keys := make([]int, 0, len(l.active))
	for key, _ := range l.active {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	inspectors := make([]Inspector, len(l.active))
	for i, index := range keys {
		inspectors[i] = l.active[index]
	}

	return inspectors
}
