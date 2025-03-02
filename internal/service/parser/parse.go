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
	Type     NodeType
	Value    rune
	Children []*Node
}

func NewNode(t NodeType, c []*Node) *Node {
	return &Node{
		Type:     t,
		Value:    '0',
		Children: c,
	}
}

func NewRuneNode(v rune) *Node {
	return &Node{
		Type:     SimpleNode,
		Value:    v,
		Children: nil,
	}
}

func (n *Node) Add(other *Node) {
	if len(n.Children) == 1 && (n.Type == GroupRefNode || n.Type == StringRefNode) {
		panic("group ref and string ref can have only one child")
	}
	n.Children = append(n.Children, other)
}

func (n *Node) GetLastChild() *Node {
	return n.Children[len(n.Children)-1]
}

func (n *Node) SetLastChild(v *Node) {
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

func (s *Service) Parse(ctx context.Context, regex string) (*Tree, error) {
	r := []rune(regex)

	st := make([]*Node, 1)
	st[0] = NewNode(SimpleNode, nil)
	tr := NewTree(st[0])

	var i int
	var brCount, grCount int
	var maxNum rune

	for i < len(r) {
		_, ok := s.allowedChars[r[i]]
		if !ok {
			return nil, errors.New("char not allowed: " + string(r[i]))
		}

		if r[i] == '*' && i-1 >= 0 && s.IsLetter(r[i-1]) {
			c := st[len(st)-1].GetLastChild()
			cS := make([]*Node, 1)
			cS[0] = c
			n := NewNode(RepeatableNode, cS)
			st[len(st)-1].SetLastChild(n)
		} else if r[i] == '|' && i-1 >= 0 && i+1 <= len(r)-1 {
			if st[len(st)-1].Type == BranchNode {
				st[len(st)-2].Add(st[len(st)-1])
				st = st[:len(st)-1]
				st = append(st, NewNode(BranchNode, nil))
			} else {
				st[len(st)-1].Type = AlternativeNode
				n := NewNode(BranchNode, st[len(st)-1].Children)
				st[len(st)-1].Children = make([]*Node, 1)
				st[len(st)-1].Children[0] = n
				st = append(st, NewNode(BranchNode, nil))
			}
		} else if i < len(r)-3 && r[i] == '(' && r[i+1] == '?' && s.IsDigit(r[i+2]) && r[i+3] == ')' {
			if r[i+2] > maxNum {
				maxNum = r[i+2]
			}
			i += 4
		} else if i < len(r)-3 && r[i] == '(' && r[i+1] == '\\' && s.IsDigit(r[i+2]) && r[i+3] == ')' {
			if int(r[i+2]-'0') > grCount {
				return nil, errors.New(
					fmt.Sprintf("cannot use str ref: groups < x (%v < %c)", grCount, r[i+2]),
				)
			}
			i += 4
		} else if i < len(r)-1 && r[i] == '\\' && s.IsDigit(r[i+1]) {
			if int(r[i+1]-'0') > grCount {
				return nil, errors.New(
					fmt.Sprintf("cannot use str ref: groups < x (%v < %c)", grCount, r[i+1]),
				)
			}
			i += 2
		} else if r[i] == '(' {
			n := NewNode(GroupNode, nil)
			st[len(st)-1].Add(n)
			st = append(st, n)
			brCount++
		} else if r[i] == ')' && brCount > 0 {
			brCount--
			grCount++
			if st[len(st)-1].Type == BranchNode {
				st[len(st)-2].Add(st[len(st)-1])
				st = st[:len(st)-1]
			}
			tr.Groups[grCount] = st[len(st)-1]
			st = st[:len(st)-1]
		} else if s.IsLetter(r[i]) {
			n := NewRuneNode(r[i])
			st[len(st)-1].Add(n)
		} else {
			return nil, errors.New(fmt.Sprintf("cannot use symbol %c on position %v", r[i], i))
		}

		i++
	}

	if brCount != 0 || grCount > 9 || int(maxNum-'0') > grCount {
		return nil, errors.New("additional constraints were not met")
	}

	return tr, nil
}
