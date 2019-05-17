package apple2

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

type persistent interface {
	save(w io.Writer)
	load(r io.Reader)
}

type persistance struct {
	a     *Apple2
	items []persistent
}

func newPersistance(a *Apple2) *persistance {
	var p persistance
	p.a = a
	return &p
}

func (p *persistance) register(i persistent) {
	p.items = append(p.items, i)
}

func (p *persistance) save(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	binary.Write(w, binary.BigEndian, p.a.isColor)
	binary.Write(w, binary.BigEndian, p.a.fastMode)
	binary.Write(w, binary.BigEndian, p.a.fastRequestsCounter)
	p.a.cpu.Save(w)

	for _, v := range p.items {
		v.save(w)
	}
}

func (p *persistance) load(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		// Ignore error if can't load the file
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)

	binary.Read(r, binary.BigEndian, &p.a.isColor)
	binary.Read(r, binary.BigEndian, &p.a.fastMode)
	binary.Read(r, binary.BigEndian, &p.a.fastRequestsCounter)
	p.a.cpu.Load(r)

	for _, v := range p.items {
		v.load(r)
	}
}
