package parser

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

const (
	SimpleNode      NodeType = iota // children are just nodes
	RepeatableNode                  // same as simple, but repeatable
	GroupNode                       // children are a group
	AlternativeNode                 // children are branches, subset of GroupNode
	BranchNode                      // branch of alternative node
	GroupRefNode                    // only one child as ref
	StringRefNode                   // only one child as ref
)

type NodeType int

type Status int

const (
	Unknown Status = iota
	VisitingChildren
	VisitedChildren
)

type Node struct {
	Value     rune
	IsIgnored bool
	GroupNum  int
	RefToNum  int
	status    Status
	Type      NodeType
	Parent    *Node
	Children  []*Node
}

func NewNode(t NodeType, p *Node, c []*Node) *Node {
	return &Node{
		Value:    '0',
		RefToNum: -1,
		Type:     t,
		Parent:   p,
		Children: c,
	}
}

func NewRuneNode(v rune, p *Node) *Node {
	return &Node{
		Value:    v,
		RefToNum: -1,
		Type:     SimpleNode,
		Parent:   p,
		Children: nil,
	}
}

func NewRefNode(t NodeType, n int, p *Node) *Node {
	return &Node{
		Value:    '0',
		RefToNum: n,
		Type:     t,
		Parent:   p,
		Children: nil,
	}
}

func (n *Node) Add(other *Node) {
	if len(n.Children) == 1 && (n.Type == GroupRefNode || n.Type == StringRefNode) {
		panic("group ref and string ref can have only one child")
	}
	other.Parent = n
	n.Children = append(n.Children, other)
}

func (n *Node) GetLastChild() *Node {
	return n.Children[len(n.Children)-1]
}

func (n *Node) SetLastChild(v *Node) {
	v.Parent = n
	n.Children[len(n.Children)-1] = v
}

type Tree struct {
	Groups map[int]*Node
	Root   *Node
}

func NewTree(r *Node) *Tree {
	return &Tree{
		Groups: make(map[int]*Node),
		Root:   r,
	}
}

func (t *Tree) CheckStringrefs() error {
	g := new(errgroup.Group)

	var f func(visited map[int]struct{}, st []*Node) error
	f = func(visited map[int]struct{}, st []*Node) error {
		for len(st) > 0 {
			el := st[len(st)-1]
			st = st[:len(st)-1]
			if el.Type == StringRefNode {
				if _, ok := visited[el.RefToNum]; !ok {
					return fmt.Errorf("string ref was not initialized: %v", el.RefToNum)
				}
			} else if el.Type == GroupNode {
				if el.status == VisitingChildren {
					visited[el.GroupNum] = struct{}{}
					el.status = VisitedChildren
				} else if el.status == Unknown {
					el.status = VisitingChildren
					st = append(st, el)
					for i := range el.Children {
						st = append(st, el.Children[len(el.Children)-1-i])
					}
				}
			} else if el.Type == RepeatableNode {
				newSt := make([]*Node, len(st))
				for i := range st {
					newSt[i] = st[i]
				}
				newVisited := make(map[int]struct{})
				for k := range visited {
					newVisited[k] = struct{}{}
				}
				g.Go(func() error {
					return f(newVisited, newSt)
				})

				newNewSt := make([]*Node, len(newSt))
				for i := range st {
					newNewSt[i] = newSt[i]
				}
				newNewSt = append(newNewSt, el.GetLastChild())
				newNewVisited := make(map[int]struct{})
				for k := range visited {
					newNewVisited[k] = struct{}{}
				}
				g.Go(func() error {
					return f(newNewVisited, newNewSt)
				})
			} else if el.Type == AlternativeNode {
				visited[el.GroupNum] = struct{}{}
				for _, cc := range el.Children {
					newSt := make([]*Node, len(st))
					for i := range st {
						newSt[i] = st[i]
					}
					newSt = append(newSt, cc)

					newVisited := make(map[int]struct{})
					for k := range visited {
						newVisited[k] = struct{}{}
					}

					g.Go(func() error {
						return f(newVisited, newSt)
					})
				}
			} else {
				for i := range el.Children {
					st = append(st, el.Children[len(el.Children)-1-i])
				}
			}
		}

		return nil
	}

	st := make([]*Node, 0)
	for i := range t.Root.Children {
		st = append(st, t.Root.Children[len(t.Root.Children)-1-i])
	}

	g.Go(func() error {
		return f(make(map[int]struct{}), st)
	})

	return g.Wait()
}

func (s *Service) Parse(ctx context.Context, regex string) (*Tree, error) {
	r := []rune(regex)

	st := make([]*Node, 1)
	st[0] = NewNode(SimpleNode, nil, nil)
	tr := NewTree(st[0])

	grIndices := make([]int, 0)

	var i int
	var brCount, grCount int
	var maxNum rune

	for i < len(r) {
		_, ok := s.allowedChars[r[i]]
		if !ok {
			return nil, errors.New("char not allowed: " + string(r[i]))
		}

		if r[i] == '*' && i-1 >= 0 && (s.IsLetter(r[i-1]) || r[i-1] == ')') {
			c := st[len(st)-1].GetLastChild()
			cS := make([]*Node, 1)
			cS[0] = c
			n := NewNode(RepeatableNode, st[len(st)-1], cS)
			c.Parent = n
			st[len(st)-1].SetLastChild(n)
			i++
		} else if r[i] == '|' && i-1 >= 0 && i+1 <= len(r)-1 {
			if st[len(st)-1].Type == BranchNode {
				st[len(st)-2].Add(st[len(st)-1])
				st = st[:len(st)-1]
				n := NewNode(BranchNode, st[len(st)-1], nil)
				st = append(st, n)
			} else {
				st[len(st)-1].Type = AlternativeNode
				n := NewNode(BranchNode, st[len(st)-1], st[len(st)-1].Children)
				st[len(st)-1].Children = make([]*Node, 1)
				st[len(st)-1].Children[0] = n
				st = append(st, NewNode(BranchNode, st[len(st)-1], nil))
			}
			i++
		} else if i < len(r)-2 && r[i] == '(' && r[i+1] == '?' && r[i+2] == ':' {
			n := NewNode(GroupNode, st[len(st)-1], nil)
			n.IsIgnored = true
			st[len(st)-1].Add(n)
			st = append(st, n)
			brCount++
			i += 3
		} else if i < len(r)-3 && r[i] == '(' && r[i+1] == '?' && s.IsDigit(r[i+2]) && r[i+3] == ')' {
			if r[i+2] > maxNum {
				maxNum = r[i+2]
			}
			n := NewRefNode(GroupRefNode, int(r[i+2]-'0'), st[len(st)-1])
			st[len(st)-1].Add(n)
			i += 4
		} else if i < len(r)-1 && r[i] == '\\' && s.IsDigit(r[i+1]) {
			grNum := int(r[i+1] - '0')
			if grNum > grCount {
				return nil, errors.New(
					fmt.Sprintf("cannot use str ref: groups < x (%v < %c)", grCount, r[i+1]),
				)
			}
			n := NewRefNode(StringRefNode, int(r[i+1]-'0'), st[len(st)-1])
			st[len(st)-1].Add(n)
			i += 2
		} else if r[i] == '(' {
			n := NewNode(GroupNode, st[len(st)-1], nil)
			st[len(st)-1].Add(n)
			st = append(st, n)
			brCount++
			grCount++
			grIndices = append(grIndices, grCount)
			n.GroupNum = grCount
			i++
		} else if r[i] == ')' && brCount > 0 {
			brCount--
			if st[len(st)-1].Type == BranchNode {
				st[len(st)-2].Add(st[len(st)-1])
				st = st[:len(st)-1]
			}
			if !st[len(st)-1].IsIgnored {
				tr.Groups[grIndices[len(grIndices)-1]] = st[len(st)-1]
				grIndices = grIndices[:len(grIndices)-1]
			}
			st = st[:len(st)-1]
			i++
		} else if s.IsLetter(r[i]) {
			n := NewRuneNode(r[i], st[len(st)-1])
			st[len(st)-1].Add(n)
			i++
		} else {
			return nil, errors.New(fmt.Sprintf("cannot use symbol %c on position %v", r[i], i))
		}
	}

	if brCount != 0 || grCount > 9 || int(maxNum-'0') > grCount {
		return nil, errors.New("additional constraints were not met")
	}

	err := tr.CheckStringrefs()
	if err != nil {
		return nil, err
	}

	return tr, nil
}
