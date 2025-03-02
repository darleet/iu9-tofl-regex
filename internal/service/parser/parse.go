package parser

import (
	"context"
	"errors"
	"fmt"
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

type Node struct {
	Value     rune
	IsIgnored bool
	RefToNum  int
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
	Groups       map[int]*Node
	StrictGroups map[int]struct{}
	Root         *Node
}

func NewTree(r *Node) *Tree {
	return &Tree{
		Groups:       make(map[int]*Node),
		StrictGroups: make(map[int]struct{}),
		Root:         r,
	}
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
			tr.StrictGroups[grNum] = struct{}{}
			i += 2
		} else if r[i] == '(' {
			n := NewNode(GroupNode, st[len(st)-1], nil)
			st[len(st)-1].Add(n)
			st = append(st, n)
			brCount++
			grCount++
			grIndices = append(grIndices, grCount)
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

	err := s.checkStrictGroups(tr)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (s *Service) checkStrictGroups(tr *Tree) error {
	for k, v := range tr.Groups {
		if _, ok := tr.StrictGroups[k]; !ok {
			continue
		}

		c := make([]*Node, 0)
		for _, cc := range v.Children {
			c = append(c, cc)
		}

		var j int
		for j < len(c) {
			if c[j].Type == GroupRefNode {
				return errors.New("strict group can't have group ref")
			}
			c = append(c, c[j].Children...)
			j++
		}

		p := v.Parent
		for p != nil {
			if p.Type == AlternativeNode {
				return errors.New("strict group can't be in alternative")
			}
			if p.Type == RepeatableNode {
				return errors.New("strict group can't be in repeatable")
			}
			p = p.Parent
		}
	}

	return nil
}
