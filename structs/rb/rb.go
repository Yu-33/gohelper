package rb

import (
	"github.com/Yu-33/gohelper/structs/bst"
	"github.com/Yu-33/gohelper/structs/container"
)

const (
	red int8 = iota
	black
)

type Element = container.Comparer

// Node used in rb tree and implements bst.Node.
type Node struct {
	element Element
	left    *Node
	right   *Node
	parent  *Node
	color   int8
}

func (n *Node) Element() Element {
	return n.element
}

func (n *Node) Left() bst.Node {
	return n.left
}

func (n *Node) Right() bst.Node {
	return n.right
}

// Tree implements data struct of red-black tree.
type Tree struct {
	root *Node
	len  int
}

// New creates an red-black Tree.
func New() *Tree {
	tr := &Tree{
		root: nil,
		len:  0,
	}
	return tr
}

// Len return number of elements.
func (tr *Tree) Len() int {
	return tr.len
}

// Search search element in RB Tree, return nil if element not exists.
func (tr *Tree) Search(element Element) Element {
	p := tr.root
	for p != nil {
		flag := element.Compare(p.element)
		if flag == -1 {
			p = p.left
		} else if flag == 1 {
			p = p.right
		} else {
			return p.element
		}
	}
	return nil
}

// Insert insert an element into RB Tree, return false if have duplicate element.
func (tr *Tree) Insert(element Element) bool {
	var n *Node
	p := tr.root
	for p != nil {
		flag := element.Compare(p.element)
		if flag == -1 {
			if p.left == nil {
				n = tr.createNode(element, p)
				p.left = n
				break
			}
			p = p.left
		} else if flag == 1 {
			if p.right == nil {
				n = tr.createNode(element, p)
				p.right = n
				break
			}
			p = p.right
		} else {
			return false
		}
	}
	if n == nil {
		n = tr.createNode(element, p)
	}

	tr.insertBalance(n)
	tr.len++

	return true
}

// Delete delete an element from RB Tree, return false if element not exists.
func (tr *Tree) Delete(element Element) Element {
	d := tr.root
	for d != nil {
		flag := element.Compare(d.element)
		if flag == -1 {
			d = d.left
		} else if flag == 1 {
			d = d.right
		} else {
			break
		}
	}
	if d == nil {
		return nil
	}

	if d.left != nil && d.right != nil {
		x := d.left
		for x.right != nil {
			x = x.right
		}

		d.element, x.element = x.element, d.element
		d = x
	}

	var c *Node

	if d.left != nil {
		c = d.left
	} else {
		c = d.right
	}

	if c != nil {
		c.parent = d.parent
	}

	if d.parent == nil {
		tr.root = c
	} else if d.parent.left == d {
		d.parent.left = c
	} else {
		d.parent.right = c
	}

	if d.color == black {
		tr.deleteBalance(c, d.parent)
	}

	tr.len--
	return d.element
}

// Iter return a Iterator, include elements: start <= k <= boundary.
// start == first node if start == nil and boundary == last node if boundary == nil.
func (tr *Tree) Iter(start Element, boundary Element) container.Iterator {
	it := bst.NewIterator(tr.root, start, boundary)
	return it
}

func (tr *Tree) insertBalance(n *Node) {
	if n.parent == nil {
		n.color = black
		tr.root = n
		return
	}
	if n.parent.color == black {
		return
	}

	var (
		p, g, u *Node
	)

	p = n.parent
	g = n.parent.parent

	if g.left == p {
		u = g.right
	} else {
		u = g.left
	}

	if u != nil && u.color == red {
		g.color = red
		p.color = black
		u.color = black
		tr.insertBalance(g)
		return
	}

	if g.left == p {
		if p.right == n {
			tr.leftRotate(p)
			p = g.left
		}
		g.color = red
		p.color = black
		tr.rightRotate(g)
	} else {
		if p.left == n {
			tr.rightRotate(p)
			p = g.right
		}
		g.color = red
		p.color = black
		tr.leftRotate(g)
	}
}

func (tr *Tree) deleteBalance(n *Node, p *Node) {
	if n != nil && n.color == red {
		n.color = black
		return
	}

	if p == nil {
		tr.root = n
		return
	}

	var s *Node

	if p.left == n {
		s = p.right
		if s.color == red {
			s.color = black
			p.color = red
			tr.leftRotate(p)
			s = p.right
		}
		if (s.left == nil || s.left.color == black) && (s.right == nil || s.right.color == black) {
			s.color = red
			tr.deleteBalance(p, p.parent)
			return
		}
		if (s.left != nil && s.left.color == red) && (s.right == nil || s.right.color == black) {
			s.color = red
			s.left.color = black
			tr.rightRotate(s)
			s = p.right
		}
		if s.right != nil && s.right.color == red {
			s.color = p.color
			p.color = black
			s.right.color = black
			tr.leftRotate(p)
		}
	} else {
		s = p.left
		if s.color == red {
			s.color = black
			p.color = red
			tr.rightRotate(p)
			s = p.left
		}
		if (s.left == nil || s.left.color == black) && (s.right == nil || s.right.color == black) {
			s.color = red
			tr.deleteBalance(p, p.parent)
			return
		}
		if (s.right != nil && s.right.color == red) && (s.left == nil || s.left.color == black) {
			s.color = red
			s.right.color = black
			tr.leftRotate(s)
			s = p.left
		}
		if s.left != nil && s.left.color == red {
			s.color = p.color
			p.color = black
			s.left.color = black
			tr.rightRotate(p)
		}
	}
}

func (tr *Tree) createNode(element Element, p *Node) *Node {
	n := new(Node)
	n.element = element
	n.color = red
	n.left = nil
	n.right = nil
	n.parent = p
	return n
}

func (tr *Tree) leftRotate(n *Node) {
	r := n.right
	if r.left != nil {
		r.left.parent = n
	}

	n.right = r.left
	r.left = n

	r.parent = n.parent
	n.parent = r

	if r.parent == nil {
		tr.root = r
	} else if r.parent.left == n {
		r.parent.left = r
	} else {
		r.parent.right = r
	}
}

func (tr *Tree) rightRotate(n *Node) {
	l := n.left
	if l.right != nil {
		l.right.parent = n
	}

	n.left = l.right
	l.right = n

	l.parent = n.parent
	n.parent = l

	if l.parent == nil {
		tr.root = l
	} else if l.parent.left == n {
		l.parent.left = l
	} else {
		l.parent.right = l
	}
}
