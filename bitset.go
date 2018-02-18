package main

const bitsize = 64

// Set ensures that the given bit is set in the BitSet.
func SetBit(s *[]uint64, i uint) {
	if len(*s) < int(i/bitsize+1) {
		r := make([]uint64, i/bitsize+1)
		copy(r, *s)
		*s = r
	}
	(*s)[i/bitsize] |= 1 << (i % bitsize)
}

// Clear ensures that the given bit is cleared (not set) in the BitSet.
func ClearBit(s *[]uint64, i uint) {
	if len(*s) >= int(i/bitsize+1) {
		(*s)[i/bitsize] &^= 1 << (i % bitsize)
	}
}

// IsSet returns true if the given bit is set, false if it is cleared.
func IsBitSet(s *[]uint64, i uint) bool {
	return (*s)[i/bitsize]&(1<<(i%bitsize)) != 0
}