// ofc
package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type OFCevolver interface {
	doOneStep(sys StressLattice) uint
}

func doOneStep(sys *StressLattice, alpha, stressNoise float32, R uint8, Ran *rand.Rand) uint {
	maxLoc := sys.FindMaxAllStress()
	currStress := sys.GetStress(maxLoc)
	sys.LoadSystem(sys.failureStress - currStress)
	currStress = sys.failureStress
	newStress := (stressNoise * (Ran.Float32()*2.0 - 1.0)) + sys.residualStress
	stressDistOut := (1.0 - alpha) * (currStress - newStress)

	//distribute stress
	sys.RangedDistStress(maxLoc, stressDistOut, newStress, R, sys.dim)
	eventSize := uint(1)

	// distribute until no more failed sites exist
	for sys.DistributeFailedSites() {
		maxLoc = sys.FindLargestDistributeFail()
		// stress change
		currStress = sys.GetStress(maxLoc)
		newStress = (stressNoise * (Ran.Float32()*2.0 - 1.0)) + sys.residualStress
		// stress to be distributed
		stressDistOut = (1.0 - alpha) * (currStress - newStress)
		sys.RangedDistStress(maxLoc, stressDistOut, newStress, R, sys.dim)
		eventSize = eventSize + 1
	}
	return eventSize
}

func pushEvent(event uint, events []uint) []uint {
	if uint(len(events)) <= event {
		neededIn := uint(uint(100) + event - uint(len(events)))
		for u := uint(0); u < neededIn; u++ {
			events = append(events, uint(0))
		}
	}
	events[event] = events[event] + 1
	return events
}

func saveData(events []uint, timeNow uint) {
	filename := "ofcdata" + strconv.FormatUint(uint64(timeNow), 10) + ".txt"

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			// other error
			os.Remove(filename)
		}
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	for u := uint(0); u < uint(len(events)); u++ {
		if events[u] > 0 {
			dataOut := "" + strconv.FormatUint(uint64(u), 10) + "     " + strconv.FormatUint(uint64(events[u]), 10) + "\n"
			fmt.Print(dataOut)
			io.WriteString(f, dataOut)
		}
	}
	f.Close()
}

func main() {
	lat := NewStressLattice(uint16(200), uint8(2))
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	currentTime := uint(0)

	start := time.Now()
	tsum := float64(0.0)

	eventSum := make([]uint, 1000, 1000)
	event := uint(0)

	for u := uint64(0); u < uint64(2500); u++ {
		start = time.Now()
		event = doOneStep(lat, 0.05, 0.125, uint8(20), generator)
		tsum = tsum + float64(time.Since(start)/time.Microsecond)
		if u%uint64(2500) == 0 {
			fmt.Println("MeanTime | ", (tsum / float64(u)))
			fmt.Println("Time | ", currentTime)
		}
		currentTime = currentTime + 1
	}

	for u := uint64(0); u < uint64(250000); u++ {
		event = doOneStep(lat, 0.05, 0.125, uint8(20), generator)
		eventSum = pushEvent(event, eventSum)
		if u%uint64(5000) == 0 {
			saveData(eventSum, currentTime)
			fmt.Println("Time | ", currentTime)
		}
		currentTime = currentTime + 1
	}
	fmt.Println("Done!")
}
