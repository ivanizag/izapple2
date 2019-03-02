package apple2

type disk2 struct {
	mmu  *memoryManager
	slot int
}

func insertCardDisk2(mmu *memoryManager, slot int) disk2 {
	var c disk2
	c.mmu = mmu
	c.slot = slot
	return c
}
