// StressLattice
package main

import (
	"../utilLat"
	"fmt"
	"github.com/deckarep/golang-set"
	"math/rand"
	"time"
)

type stressSite struct {
	i      uint64
	stress float32
}

type StressLattice struct {
	L                             uint16
	dim                           uint8
	N                             uint64
	seed                          int64
	failureStress, residualStress float32
	lattice                       []float32
	failedSites                   mapset.Set
}

func NewStressLatticeSeeded(L uint16, dim uint8, seed int64) *StressLattice {
	sys := new(StressLattice)
	sys.L = L
	sys.dim = dim
	sys.failureStress = 1.0
	sys.residualStress = 0.625
	rand.Seed(seed)
	sys.seed = seed
	sys.N = powInteger(L, dim)
	sys.lattice = make([]float32, sys.N, sys.N)
	sys.randomSysInit()
	sys.failedSites = mapset.NewSet()
	return sys
}

func NewStressLattice(L uint16, dim uint8) *StressLattice {
	sys := new(StressLattice)
	sys.L = L
	sys.dim = dim
	sys.failureStress = 1.0
	sys.residualStress = 0.625
	rand.Seed(time.Now().UTC().UnixNano())
	sys.seed = rand.Int63()
	rand.Seed(sys.seed)
	sys.N = powInteger(L, dim)
	sys.lattice = make([]float32, sys.N, sys.N)
	sys.randomSysInit()
	sys.failedSites = mapset.NewSet()
	return sys
}

func powInteger(a uint16, b uint8) uint64 {
	powerOf := uint64(1)
	for i := uint8(0); i < b; i++ {
		powerOf *= uint64(a)
	}
	return powerOf
}

func (f *StressLattice) LoadSystem(loadStress float32) {
	for i := uint64(0); i < f.N; i++ {
		f.SetStress(i, f.GetStress(i)+loadStress)
	}
}

func (f *StressLattice) SetStress(location uint64, newStress float32) {
	f.lattice[location] = newStress
}

func (f *StressLattice) GetStress(location uint64) float32 {
	return f.lattice[location]
}

func (f *StressLattice) randomSysInit() {
	for i := uint64(0); i < f.N; i++ {
		f.randomInitStress(i)
	}
}

func (f *StressLattice) randomInitStress(location uint64) {
	f.lattice[location] = (rand.Float32() * (f.failureStress - f.residualStress)) + f.residualStress
}

func (f *StressLattice) SiteFailed(location uint64) bool {
	if f.lattice[location] > f.failureStress {
		return true
	} else {
		return false
	}
	return true
}

func (f *StressLattice) RangedDistStress(loc uint64, distStress, newStress float32, r, dim uint8) bool {
	if dim > 3 {
		fmt.Println("Dimension greater than 3 not allowed.")
		return false
	}
	siteFailed := false

	// remove site from failed set if in failed set
	if f.failedSites.Contains(loc) {
		f.failedSites.Remove(loc)
	}
	f.SetStress(loc, newStress)

	// divide stress out for all neighbors in range
	interL := uint16(2*r + 1)
	ninteract := powInteger(interL, dim)
	distVal := (distStress) / (float32(ninteract))
	//fmt.Println("Dist | ", distVal)

	// calculate x y z coordinates of stress location
	i := utilLat.GetX(loc, f.L)
	j := utilLat.GetY(loc, f.L)
	k := utilLat.GetZ(loc, f.L)

	for u := uint64(0); u < (ninteract + uint64(1)); u++ {
		// calculate x y z coordinates of stress distribute location
		a := utilLat.GetX(u, interL)
		b := uint64(r)
		c := uint64(r)

		if dim > 1 {
			b = uint64(utilLat.GetY(u, interL))
		}

		if dim > 2 {
			c = uint64(utilLat.GetZ(u, interL))
		}

		// index of updated stress location
		ind := utilLat.GetIndex3D(
			utilLat.GetX(uint64(i)+uint64(f.L)-uint64(r)+uint64(a), f.L),
			utilLat.GetX(uint64(j)+uint64(f.L)-uint64(r)+uint64(b), f.L),
			utilLat.GetX(uint64(k)+uint64(f.L)-uint64(r)+uint64(c), f.L), f.L)

		// set site to new stress
		if ind != loc {
			newStress := distVal + f.GetStress(ind)
			f.SetStress(ind, newStress)
			// note if new site fails
			if f.SiteFailed(ind) {
				f.failedSites.Add(ind)
			}
		}
	}
	return siteFailed
}

func (f *StressLattice) FindLargestDistributeFail() uint64 {
	maxLoc := uint64(0)
	max := float32(-1.0)
	val := max
	for fsite := range f.failedSites {
		val = f.GetStress(fsite.(uint64))
		if val > max {
			max = val
			maxLoc = fsite.(uint64)
		}
	}
	return maxLoc
}

func (f *StressLattice) DistributeFailedSites() bool {
	avalanche := false
	if f.failedSites.Size() > 0 {
		avalanche = true
	}
	return avalanche
}

func (f *StressLattice) AnySitesFailed() bool {
	failed := false
	for i := uint64(0); i < f.N; i++ {
		if f.SiteFailed(i) {
			failed = true
		}
	}
	return failed
}

func (f *StressLattice) FindMaxAllStress() uint64 {
	maxLoc := uint64(0)
	max := float32(-1.0)
	val := max
	for i := uint64(0); i < f.N; i++ {
		val = f.GetStress(i)
		if val > max {
			max = val
			maxLoc = i
		}
	}
	return maxLoc
}

