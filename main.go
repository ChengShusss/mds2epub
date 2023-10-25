package main

import (
	"fmt"
	"os"
)

type OperationFunc func()

type Operation struct {
	Name string
	Desp string
	Func OperationFunc
}

type OperationSet struct {
	set map[string]*Operation
}

func (s *OperationSet) AddOperation(name, desp string, f OperationFunc) {
	s.set[name] = &Operation{name, desp, f}
}

func (s *OperationSet) ParseAndHandle(operation string) {
	op, ok := s.set[operation]
	if !ok {
		s.PrintInfo()
		return
	}
	op.Func()
}

var operationSet = OperationSet{
	set: map[string]*Operation{},
}

func (s *OperationSet) PrintInfo() {
	maxlen := 0
	for _, op := range s.set {
		if len(op.Name) > maxlen {
			maxlen = len(op.Name) + 5
		}
	}

	for _, op := range s.set {
		format := fmt.Sprintf("%%-%ds%%s\n", maxlen)
		fmt.Printf(format, op.Name, op.Desp)
	}

}

func main() {

	operationSet.AddOperation(
		"pack",
		"pack contents under specific folder into epub file",
		Pack)

	operationSet.AddOperation(
		"trans",
		"trans markdown files into epub file",
		Trans)

	if len(os.Args) == 1 {
		operationSet.PrintInfo()
		return
	}

	action := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)
	operationSet.ParseAndHandle(action)
}
