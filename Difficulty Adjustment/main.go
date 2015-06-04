package main

import (
	"bytes"
	"fmt"
	"math/big"
	"os"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/types"
)

// adjustTarget returns a target after it has been adjusted.
func adjustTarget(initial types.Target, adjustment *big.Rat) types.Target {
	adjustedRatTarget := new(big.Rat).Mul(initial.Rat(), adjustment)
	return types.RatToTarget(adjustedRatTarget)
}

// targetAdjustment0 adjusts the target linearly according to time and applies
// a clamp above 1%.
func targetAdjustment0(timePassed, expectedTimePassed int64) *big.Rat {
	base := big.NewRat(timePassed, expectedTimePassed)
	if base.Cmp(big.NewRat(1001, 1000)) > 0 {
		return big.NewRat(1001, 1000)
	} else if base.Cmp(big.NewRat(999, 1000)) < 0 {
		return big.NewRat(999, 1000)
	}
	return base
}

func targetAdjustment2(timePassed, expectedTimePassed int64) *big.Rat {
	base := big.NewRat(timePassed, expectedTimePassed)
	if base.Cmp(big.NewRat(4, 1)) > 0 {
		return big.NewRat(4, 1)
	} else if base.Cmp(big.NewRat(1, 4)) < 0 {
		return big.NewRat(1, 4)
	}
	return base
}

// simulation0 simulates a bunch of blocks using targetAdjustment0. The
// difficulty is printed (as a bar) after each difficulty adjustment.
func simulation0() {
	var target types.Target
	target[1] = 255
	window := int64(1000)
	rate := int64(750)
	blocks := int64(0)
	time := int64(0)
	max := int64(0)
	blockMem := make(map[int64]int64)
	for blocks < 25000 {
		blocks++

		// Grind nonces until the target is met, finding a block.
		hash := crypto.HashObject(time)
		for bytes.Compare(target[:], hash[:]) <= 0 {
			time++
			hash = crypto.HashObject(time)
		}
		blockMem[blocks] = time

		// Adjust the target
		var adjustment *big.Rat
		if blocks%1 == 0 {
			if blocks < window {
				adjustment = targetAdjustment0(time, blocks*rate)
			} else {
				reference, _ := blockMem[blocks-window]
				adjustment = targetAdjustment0(time-reference, window*rate)
			}
			target = adjustTarget(target, adjustment)
		}

		// Print the time + difficulty
		if blocks%1 == 0 {
			difficulty := new(big.Int).Div(types.RootDepth.Int(), target.Int())
			// fmt.Printf("Block %v, %v, %v\n", blocks, time, difficulty)
			d := difficulty.Int64()
			if d > max {
				max = d
			}
			a := int64(0)
			for a < d {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("+")
				}
			}
			for a < max {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("-")
				}
			}
			fmt.Println()
		}
		time++
	}
	fmt.Printf("Expected: %v, Actual: %v\n", blocks*rate, time)
}

func targetAdjustment1(timePassed, expectedTimePassed int64) *big.Rat {
	base := big.NewRat(timePassed, expectedTimePassed)
	if base.Cmp(big.NewRat(25, 10)) > 0 {
		return big.NewRat(25, 10)
	} else if base.Cmp(big.NewRat(10, 25)) < 0 {
		return big.NewRat(10, 25)
	}
	return base
}

func simulation1() {
	var target types.Target
	target[0] = 4
	window := int64(1000)
	rate := int64(750)
	blocks := int64(0)
	time := int64(0)
	max := int64(0)
	freq := int64(500)
	blockMem := make(map[int64]int64)
	for blocks < 125000 {
		blocks++

		// Grind nonces until the target is met, finding a block.
		hash := crypto.HashObject(time)
		for bytes.Compare(target[:], hash[:]) <= 0 {
			time++
			hash = crypto.HashObject(time)
		}
		blockMem[blocks] = time

		// Adjust the target
		var adjustment *big.Rat
		if blocks%freq == 0 {
			if blocks < window {
				adjustment = targetAdjustment1(time, blocks*rate)
			} else {
				reference, _ := blockMem[blocks-window]
				fmt.Println(blocks)
				fmt.Println(blocks-window)
				adjustment = targetAdjustment1(time-reference, window*rate)
			}
			target = adjustTarget(target, adjustment)
		}

		// Print the time + difficulty
		if blocks%freq == 0 {
			difficulty := new(big.Int).Div(types.RootDepth.Int(), target.Int())
			// fmt.Printf("Block %v, %v, %v\n", blocks, time, difficulty)
			d := difficulty.Int64()
			if d > max {
				max = d
			}
			a := int64(0)
			for a < d {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("+")
				}
			}
			for a < max {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("-")
				}
			}
			fmt.Println()
		}
		time++
	}
	fmt.Printf("Expected: %v, Actual: %v\n", blocks*rate, time)
}

func simulation2() {
	var target types.Target
	target[1] = 255
	window := int64(2016)
	rate := int64(750)
	blocks := int64(0)
	time := int64(0)
	max := int64(0)
	blockMem := make(map[int64]int64)
	for blocks < 125000 {
		blocks++

		// Grind nonces until the target is met, finding a block.
		hash := crypto.HashObject(time)
		for bytes.Compare(target[:], hash[:]) <= 0 {
			time++
			hash = crypto.HashObject(time)
		}
		blockMem[blocks] = time

		// Adjust the target
		var adjustment *big.Rat
		if blocks%2016 == 0 {
			if blocks < window {
				adjustment = targetAdjustment2(time, blocks*rate)
			} else {
				reference, _ := blockMem[blocks-window]
				adjustment = targetAdjustment2(time-reference, window*rate)
			}
			target = adjustTarget(target, adjustment)
		}

		// Print the time + difficulty
		if blocks%2016 == 0 {
			difficulty := new(big.Int).Div(types.RootDepth.Int(), target.Int())
			// fmt.Printf("Block %v, %v, %v\n", blocks, time, difficulty)
			d := difficulty.Int64()
			if d > max {
				max = d
			}
			a := int64(0)
			for a < d {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("+")
				}
			}
			for a < max {
				a += 10
				if a < rate && a+20 > rate {
					fmt.Print("|")
				} else {
					fmt.Print("-")
				}
			}
			fmt.Println()
		}
		time++
	}
	fmt.Printf("Expected: %v, Actual: %v\n", blocks*rate, time)
}

func main() {
	if os.Args[1] == "sia-original" {
		simulation0()
	}
	if os.Args[1] == "sia-new" {
		simulation1()
	}
	if os.Args[1] == "bitcoin" {
		simulation2()
	}
}
