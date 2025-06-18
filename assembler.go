package reassembler

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

type Reassembler[T any, S constraints.Integer] struct {
	name string

	sequenceFunc func(T) S
	dropCallback func(T, *Issue[S])

	buffer  sync.Map
	nextSeq S

	in  chan T
	out chan T
}

func (r *Reassembler[T, S]) In(pkg T) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("assembler '%s': panic occurred while sending package: %v", r.name, re)
		}
	}()
	r.in <- pkg
	return nil
}

func (r *Reassembler[T, S]) Out() (T, bool) {
	pkg, ok := <-r.out
	return pkg, ok
}

func (r *Reassembler[T, S]) handlePackage(pkg T) {
	var issue *Issue[S]

	defer func() {
		if re := recover(); re != nil {
			issue = NewIssue(IssueTypePanicOccurred, r.nextSeq, fmt.Errorf("assembler '%s': panic occurred while processing package: %v", r.name, re))
		}

		if issue != nil && r.dropCallback != nil {
			r.dropCallback(pkg, issue)
		}
	}()

	seq := r.sequenceFunc(pkg)

	if seq < r.nextSeq {
		issue = NewIssue(IssueTypePackageLessThanNextSeq, r.nextSeq, fmt.Errorf("assembler '%s': package sequence %v is less than next sequence %v", r.name, seq, r.nextSeq))
	} else if seq > r.nextSeq {
		r.buffer.Store(seq, pkg)
	} else {
		//fmt.Println("Write", seq, pkg)
		r.out <- pkg
		for {
			r.nextSeq += 1
			if nextPkg, ok := r.buffer.Load(r.nextSeq); ok {
				r.out <- nextPkg.(T)
				//fmt.Println("Find & Write", r.nextSeq, nextPkg)
				r.buffer.Delete(r.nextSeq)
			} else {
				break
			}
		}
	}
}

func (r *Reassembler[T, S]) Start(ctx context.Context, begin S) {
	r.nextSeq = begin
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("[*]Assembler '%s': context done, stopping...\n", r.name)
				close(r.out)
				close(r.in)
				return
			case pkg := <-r.in:
				r.handlePackage(pkg)
			}
		}
	}()
}

func (r *Reassembler[T, S]) OnDrop(callback func(T, *Issue[S])) {
	r.dropCallback = callback
}

func NewReassembler[T any, S constraints.Integer](name string, sequenceFunc func(T) S) (*Reassembler[T, S], error) {
	if sequenceFunc == nil {
		return nil, fmt.Errorf("assembler '%s': sequence function is not set", name)
	}

	return &Reassembler[T, S]{
		name:         name,
		sequenceFunc: sequenceFunc,
		buffer:       sync.Map{},
		in:           make(chan T),
		out:          make(chan T),
		dropCallback: func(t T, i *Issue[S]) {
			fmt.Printf("[!]Assembler '%s': dropped package due to issue: %v\n", name, i.Err)
		},
	}, nil
}
