package reassembler

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type Package struct {
	Sequence uint32
	Data     byte
}

type Packages []Package

func (ps Packages) Len() int {
	return len(ps)
}

func (ps Packages) Data() []byte {
	data := make([]byte, len(ps))
	for i, p := range ps {
		data[i] = p.Data
	}
	return data
}

var F func(pkg Package) uint32

func TestNewReassembler(t *testing.T) {
	rawData := []byte("Hello, Reassembler!")

	assembler := NewReassembler("order", F)

	ctx, cancel := context.WithCancel(context.Background())
	assembler.Start(ctx, 0)

	// 创建有序的包
	packages := make(Packages, len(rawData))
	for i := range rawData {
		packages[i] = Package{
			Sequence: uint32(i),
			Data:     rawData[i],
		}
	}

	// 打乱包的顺序
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(packages), func(i, j int) {
		packages[i], packages[j] = packages[j], packages[i]
	})

	fmt.Println(string(packages.Data()))

	go func() {
		in := assembler.In()
		for _, pkg := range packages {
			in <- pkg
		}
		cancel()
	}()

	received := bytes.NewBuffer(nil)
	out := assembler.Out()
	for pkg := range out {
		//fmt.Printf("- Received package: Sequence=%d, Data=%c\n", pkg.Sequence, pkg.Data)
		received.WriteByte(pkg.Data)

	}
	fmt.Printf("Received data: %s\n", received.String())

}
