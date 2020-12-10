package bst

import (
	"reflect"

	"github.com/Yu-33/gohelper/structs/stack"
)

// SearchRange get the specified range of data.
// The range is start <= x < boundary, and we allowed the start or boundary is nil.
func SearchRange(root Node, start Key, boundary Key) []KV {
	if root == nil {
		return nil
	}

	var result []KV

	s := stack.Default()
	p := root
	for !s.Empty() || !reflect.ValueOf(p).IsNil() {
		if !reflect.ValueOf(p).IsNil() {
			if start != nil && p.Key().Compare(start) == -1 {
				p = p.Right()
				continue
			}
			if boundary != nil && p.Key().Compare(boundary) != -1 {
				p = p.Left()
				continue
			}
			s.Push(p)
			p = p.Left()
		} else {
			p = s.Pop().(Node)
			result = append(result, p)
			p = p.Right()
		}
	}

	return result
}

// SearchLastLT search for the last node that less than the 'key'.
func SearchLastLT(root Node, key Key) KV {
	if root == nil || key == nil {
		return nil
	}

	var n Node

	p := root
	for !reflect.ValueOf(p).IsNil() {
		flag := key.Compare(p.Key())
		if flag == 1 {
			n = p
			p = p.Right()
		} else {
			p = p.Left()
		}
	}

	if n != nil {
		return n
	}

	return nil
}

// SearchLastLE search for the last node that less than or equal to the 'key'.
func SearchLastLE(root Node, key Key) KV {
	if root == nil || key == nil {
		return nil
	}

	var n Node

	p := root
	for !reflect.ValueOf(p).IsNil() {
		flag := key.Compare(p.Key())
		if flag == 1 {
			n = p
			p = p.Right()
		} else if flag == -1 {
			p = p.Left()
		} else {
			n = p
			break
		}
	}

	if n != nil {
		return n
	}

	return nil
}

// SearchFirstGT search for the first node that greater than to the 'key'.
func SearchFirstGT(root Node, key Key) KV {
	if root == nil || key == nil {
		return nil
	}

	var n Node

	p := root
	for !reflect.ValueOf(p).IsNil() {
		flag := key.Compare(p.Key())
		if flag == -1 {
			n = p
			p = p.Left()
		} else {
			p = p.Right()
		}
	}

	if n != nil {
		return n
	}

	return nil
}

// SearchFirstGE search for the first node that greater than or equal to the 'key'.
func SearchFirstGE(root Node, key Key) KV {
	if root == nil || key == nil {
		return nil
	}

	var n Node

	p := root
	for !reflect.ValueOf(p).IsNil() {
		flag := key.Compare(p.Key())
		if flag == -1 {
			n = p
			p = p.Left()
		} else if flag == 1 {
			p = p.Right()
		} else {
			n = p
			break
		}
	}

	if n != nil {
		return n
	}

	return nil
}
