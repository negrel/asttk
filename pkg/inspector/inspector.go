package inspector

import (
	"go/ast"
	"log"
)

type Inspector func(node ast.Node) bool

// Lead is the Inspector chief that manage the inspection.
type Lead struct {
	active   []Inspector
	depth    int
	inactive map[int]map[int]Inspector
}

// Lieutenant define an Inspector that manage his own Inspectors.
func Lieutenant(Inspectors ...Inspector) Inspector {
	return New(Inspectors...).inspect
}

// New return an Inspector Lead.
func New(inspectors ...Inspector) *Lead {
	length := len(inspectors)
	ii := make([]Inspector, length, length+1)
	for i, inspector := range inspectors {
		ii[i] = inspector
	}

	return &Lead{
		active:   ii,
		depth:    0,
		inactive: make(map[int]map[int]Inspector),
	}
}

func (l *Lead) Inspect(node ast.Node) {
	l.depth = 0
	ast.Inspect(node, l.inspect)
}

func (l *Lead) inspect(node ast.Node) bool {
	if node == nil {
		defer func() {
			l.depth--
			l.recoverStoppedAt(l.depth)
		}()
	} else {
		defer func() {
			l.depth++
		}()
	}

	// log.Println(l.active)
	i := -1
	for index, inspector := range l.active {
		i++

		ok := inspector(node)
		if !ok {
			l.stopAt(l.depth, i, index)
			i--
		}
	}

	if len(l.active) == 0 {
		l.recoverStoppedAt(l.depth)
		return false
	}

	return true
}

func (l *Lead) recoverStoppedAt(depth int) {
	if _, ok := l.inactive[l.depth]; !ok {
		return
	}

	for index, inspector := range l.inactive[depth] {
		log.Println(len(l.active), l.active, depth, index, l.inactive[depth])
		if length := len(l.active); length == 0 || length <= index {
			l.active = append(l.active, inspector)
		} else {
			tmp := l.active[index+1:]
			l.active = append(l.active, inspector)
			l.active = append(l.active, tmp...)
		}
		// if length := len(l.active); length == 0 || length == index {
		// 	l.active = append(l.active, inspector)
		// } else {
		// 	log.Println(len(l.active), index, l.active)
		// 	last := length - 1
		// 	l.active = append(l.active, l.active[last])
		// 	copy(l.active[index+1:], l.active[index:last])
		// 	l.active[index] = inspector
		// }
	}

	delete(l.inactive, depth)
}

func (l *Lead) stopAt(depth, i, index int) {
	log.Println("disabling", i, l.active[i], "at depth", depth)

	if l.inactive[depth] == nil {
		l.inactive[depth] = make(map[int]Inspector)
	}

	l.inactive[depth][index] = l.active[i]
	l.active = append(l.active[:i], l.active[i+1:]...)
}
