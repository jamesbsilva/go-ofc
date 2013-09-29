// LatticeIndexed
package utilLat

import (
	"fmt"
)

type LatticeIndexer interface {
	GetIndex2D(x, y, L uint16) uint64
	GetIndex3D(x, y, z, L uint16) uint64
	GetX(i uint64, L uint16) uint16
	GetY(i uint64, L uint16) uint16
	GetZ(i uint64, L uint16) uint16
}

func GetX(i uint64, L uint16) uint16 {
	return uint16(i % uint64(L))
}

func GetY(i uint64, L uint16) uint16 {
	return uint16((i / (uint64(L))) % uint64(L))
}

func GetZ(i uint64, L uint16) uint16 {
	return uint16((i / (uint64(L) * uint64(L))) % uint64(L))
}

func GetIndex2D(x, y, L uint16) uint64 {
	return (uint64(x) + (uint64(y) * uint64(L)))
}

func GetIndex3D(x, y, z, L uint16) uint64 {
	return (uint64(x) + (uint64(y) * uint64(L)) + (uint64(z) * uint64(L) * uint64(L)))
}

func main() {
	fmt.Println("Indexer Test!")
	fmt.Println("Test | ", GetZ(21000, 100))
}
