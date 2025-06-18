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

func TestNewReassembler(t *testing.T) {
	rawData := []byte("Hello, Assembler!")

	assembler, err := NewReassembler("order", func(p Package) uint32 {
		return p.Sequence
	})

	if err != nil {
		fmt.Println("Error creating reassembler:", err)
		return
	}
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
		for _, pkg := range packages {
			if err := assembler.In(pkg); err != nil {
				fmt.Printf("- Error sending package: %v\n", err)
				break
			}
		}
		cancel()
	}()

	received := bytes.NewBuffer(nil)
	for pkg, ok := assembler.Out(); ok; pkg, ok = assembler.Out() {
		//fmt.Printf("- Received package: Sequence=%d, Data=%c\n", pkg.Sequence, pkg.Data)
		received.WriteByte(pkg.Data)

	}
	fmt.Printf("Received data: %s\n", received.String())

}
