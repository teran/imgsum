package arw

import (
	"fmt"
	"log"
)

func readbits(src byte, start byte, len byte) byte {
	return (src >> start) & ((1 << len)-1)
}

func read(row []byte, offset byte, size byte) uint16 {
	var result uint16

	for cur := offset/8; size > 0; cur++ {
		//cur: the current byte we are reading from.
		// offset: position relative to the highest bit we want the right side of
		// size: the amount of bits to read to the right of offset
		// Therefore, starting point seen form the lowest bit is offset + size
		// if offset + size is biggest than 8 the value will be right padded with zeroes
		// where it drops below the 0th bit
		start := offset%8
		if invertedStart := 8 - int((start+size));invertedStart < 0 {
			panic("Failed hard family")
		} else {
			result += uint16(readbits(row[cur],byte(invertedStart),size))
			size -=size
			log.Printf("%08b %08b %v %v\n",result,row[cur],invertedStart,size)
		}
	}
	return result
}


func readEvenGB(row []byte) string{
	var max,min uint16
	var maxOffset, minOffset uint8
	var deltas [14]uint8

	var offset byte = 0

	var size byte = 11
	max = uint16(read(row,offset,size)) //11
	offset+=size

	min = uint16(read(row,offset,size)) //22
	offset+=size

	size = 4
	maxOffset = uint8(read(row,offset,size)) //26
	offset += size

	minOffset = uint8(read(row,offset,size)) //30
	offset += size

	size = 7
	for i := range deltas {
		deltas[i] = uint8(read(row,offset,size))
		offset += size
	}

	var ret string
	ret += fmt.Sprintf("Colours interpreted as bits:\n%011b\n%011b\n%04b\n%04b\n", max, min, maxOffset, minOffset)
	for _, delta := range deltas {
		ret += fmt.Sprintf("%07b\n", delta)
	}
	ret+= fmt.Sprintf("Final offset in bits: %v\n",offset)
	return ret
}
