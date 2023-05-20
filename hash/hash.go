package hash

import "fmt"

// Single is a quick-and-dirty approximation of a JOAAT hash, returned as a lower case hexadecimal string.
func Single(original string) (result string) {
	hash := uint32(0)

	for _, b := range original {
		hash += uint32(b)
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15

	return fmt.Sprintf("%08x", hash)
}

// Mapped takes any number of strings, and returns a map with the hashes as keys and the originals as values.
// Intended use is giving it a list of entry names, making it possible to look up what actual entry is referred to by a hash later.
func Mapped(originals ...string) (result map[string]string) {
	result = make(map[string]string)
	for _, key := range originals {
		result[Single(key)] = key
	}
	return result
}
