package main

import (
	"fmt"
	"math/big"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/types"
)

func main() {
	// This simulation assumes that miners can create blocks of any size.

	// Topology: 6 miners, fully connected.
	//
	// Miners are A, B, C, D, E, F
	//
	// A-B is a 40gbps connection
	// B-C is a 40gpbs connection
	// B-C is a 40gbps connection
	//
	// All other connections are 1gbps.
	//
	// A, B, C all have 25% hashpower each. D and E have 10%, F has 5%.
	//
	// All miners are mining 'noraml' blocks except for miner A. Normal blocks
	// propagate around the network instantly. Miner A is creating
	// intentionally slow propagating blocks - 5GB blocks. All transactions are
	// new to the network, so the whole block must be downloaded after it is
	// discovered. (40 seconds to travel over a normal connection). When 'A'
	// makes a block, the block is sent immediately to B (taking 2 seconds),
	// then A stops doing any work at all. B and C work together to get the
	// block to the rest of the network.
	//
	// Propagation time of the 5GB block:
	// A -> B && B -> C: 1 second (1 total)
	// B -> D && C -> F: 40 seconds (41 total)
	// D -> E:           40 seconds (81 total)

	// Each loop represents 1 second on the network. To keep the simulation
	// fast, 100% of the network is 20 hashes per second.
	target := types.RootDepth.MulDifficulty(big.NewRat(12000, 1))

	// Demonstrate that the correct target has been chosen. Could be moved to a
	// unit test.
	/*
		{
			// The target is supposed to be 12,000 hashes per block. Perform enough
			// hashes to mine 250 blocks.
			blocksFound := 0
			for i := 0; i < 12e3 * 250; i++ {
				result := types.Target(crypto.HashObject(i))
				if result.Cmp(target) <= 0 {
					blocksFound++
				}
			}

			// Result is, remarkably, exactly 250 blocks found.
			fmt.Println(blocksFound)
		}
	*/

	// Demonstrate that the basic simulation is sane - A is being friendly in
	// this case. Could be moved to a unit test.
	/*
		{
			// Each iteration of the loop simulates 1 second on the network.
			var blocksA, blocksB, blocksC, blocksD, blocksE, blocksF int
			for i := 0; i < 600 * 500; i++ {
				// An int indicates how many blocks each miner found in this
				// second.
				var A, B, C, D, E, F int

				// A, B, C, all do 5 hashes.
				for j := 0; j < 5; j++ {
					resultA := types.Target(crypto.HashAll("A", i, j))
					resultB := types.Target(crypto.HashAll("B", i, j))
					resultC := types.Target(crypto.HashAll("C", i, j))

					if resultA.Cmp(target) <= 0 {
						A++
					}
					if resultB.Cmp(target) <= 0 {
						B++
					}
					if resultC.Cmp(target) <= 0 {
						C++
					}
				}

				// D, E do 2 hashes.
				for j := 0; j < 2; j++ {
					resultD := types.Target(crypto.HashAll("D", i, j))
					resultE := types.Target(crypto.HashAll("E", i, j))

					if resultD.Cmp(target) <= 0 {
						D++
					}
					if resultE.Cmp(target) <= 0 {
						E++
					}
				}

				// F does 1 hash.
				resultF := types.Target(crypto.HashAll("F", i))
				if resultF.Cmp(target) <= 0 {
					F++
				}

				// Process the mining. If multiple miners find blocks at the
				// same time, assume the blocks stack on top of eachother.
				blocksA += A
				blocksB += B
				blocksC += C
				blocksD += D
				blocksE += E
				blocksF += F
				A = 0
				B = 0
				C = 0
				D = 0
				E = 0
				F = 0
			}
			fmt.Println(blocksA, blocksB, blocksC, blocksD, blocksE, blocksF)
			// Expected output is: '125, 125, 125, 50, 50, 25'
			// Actual output is:   '132, 134, 122, 44, 46, 23'
			//
			// Simulation working as expected.
		}
	*/

	// Each iteration of the loop simulates 1 second on the network.
	var blocksA, blocksB, blocksC, blocksD, blocksE, blocksF int
	var stalesA, stalesB, stalesC, stalesD, stalesE, stalesF int
	var heavyA, heavyB, heavyC, heavyD, heavyE int
	var compD, compE, compF int
	var seenDE bool
	progressA := -1 // indicates which nodes have seen the most recent 'A' block.
	heavyDepth := 0 // indicates how many blocks have been built on top of 'A's most recent block.
	compDepth := 0  // indicates how many blocks have been built that compete with 'A'.
	for i := 0; i < 600*50000; i++ {
		if i%(600*500) == 0 {
			print(i / (600 * 500))
			println("% complete")
		}
		// An int indicates how many blocks each miner found in this second.
		var A, B, C, D, E, F int

		// A, B, C, all do 5 hashes.
		for j := 0; j < 5; j++ {
			resultA := types.Target(crypto.HashAll("A", i, j))
			resultB := types.Target(crypto.HashAll("B", i, j))
			resultC := types.Target(crypto.HashAll("C", i, j))

			if resultA.Cmp(target) <= 0 {
				A++
			}
			if resultB.Cmp(target) <= 0 {
				B++
			}
			if resultC.Cmp(target) <= 0 {
				C++
			}
		}

		// D, E do 2 hashes.
		for j := 0; j < 2; j++ {
			resultD := types.Target(crypto.HashAll("D", i, j))
			resultE := types.Target(crypto.HashAll("E", i, j))

			if resultD.Cmp(target) <= 0 {
				D++
			}
			if resultE.Cmp(target) <= 0 {
				E++
			}
		}

		// F does 1 hash.
		resultF := types.Target(crypto.HashAll("F", i))
		if resultF.Cmp(target) <= 0 {
			F++
		}

		// Process the mining. If multiple miners find blocks at the same time,
		// assume the blocks stack on top of eachother. The exception is 'A',
		// who will create blocks that are slow to propagate. If 'A' already
		// has a block that is not fully propagated, additional blocks produced
		// by 'A' will propagate instantly.
		//
		// Simulation is not perfect, but closely approximates the desired
		// network.
		if progressA == -1 && A > 0 {
			// A has found a block, and is not currently building off of a
			// heavy block, so A will create a heavy block.
			progressA = 0
		}
		// Any heavy block in progress has 'A' blocks added to it.
		heavyDepth += A
		heavyA += A
		if progressA == -1 {
			// There is no heavy block from A, all blocks found by other miners
			// extend the longest chain with no risk of stales.
			blocksB += B
			blocksC += C
			blocksD += D
			blocksE += E
			blocksF += F
		} else if progressA == 0 {
			// A heavy block from 'A' has been seen by nobody. If another block
			// has been found, 'A' will just give up and accept a stale. It is
			// not optimal to try mining a heavy block with only 25% hashpower.
			if B+C+D+E+F > 0 {
				println("A heavy block was thrown away for being inviable")
				stalesA += heavyDepth
				progressA = -1
				blocksB += B
				blocksC += C
				blocksD += D
				blocksE += E
				blocksF += F
			}
		} else if progressA < 41 {
			// A heavy block from 'A' has been seen by 'B' and 'C', but not by
			// the slower miners.
			heavyDepth += B + C
			heavyB += B
			heavyC += C

			// Remaining blocks will extend the competing chain.
			if D+E+F > 0 {
				compDepth += D + E + F
				compD += D
				compE += E
				compF += F
				seenDE = true
			}
		} else if progressA < 81 {
			// The block has propagated everywhere except to 'F'.
			heavyDepth += B + C
			heavyB += B
			heavyC += C

			// D, E will be mining on the heavy block if the heavy block at any
			// point gets ahead of the light chain.
			if heavyDepth > compDepth {
				seenDE = false
			}

			// D + E will always have the same blockchain due to this topology.
			if !seenDE {
				heavyDepth += D + E
				heavyD += D
				heavyE += E
			} else {
				compDepth += D + E
				compD += D
				compE += E
			}

			// F will extend the competing chain.
			if F > 0 {
				compDepth += F
				compF += F
			}
		} else {
			// The heavy block has propagated everywhere, but the competing
			// chain is at the same height.
			heavyDepth += B + C
			heavyB += B
			heavyC += C

			if !seenDE {
				heavyDepth += D + E
				heavyD += D
				heavyE += E
			} else {
				compDepth += D + E
				compD += D
				compE += E
			}

			// F will extend the competing chain.
			if F > 0 {
				compDepth += F
				compF += F
			}
		}
		// If at any point, the competing chain gets ahead of the heavy chain,
		// the heavy chain will be dropped entirely.
		if compDepth > heavyDepth {
			println("A competition chain has defeated a heavy chain")
			progressA = -1
			stalesA += heavyA
			stalesB += heavyB
			stalesC += heavyC
			stalesD += heavyD
			stalesE += heavyE
			blocksD += compD
			blocksE += compE
			blocksF += compF

			heavyA = 0
			heavyB = 0
			heavyC = 0
			heavyD = 0
			heavyE = 0
			compD = 0
			compE = 0
			compF = 0
			compDepth = 0
			heavyDepth = 0
		}
		if progressA != -1 && progressA < 82 {
			progressA++
		} else if progressA <= 82 && heavyDepth > compDepth {
			if compDepth > 0 {
				println("A heavy chain has defeated a competition chain")
			}
			progressA = -1
			blocksA += heavyA
			blocksB += heavyB
			blocksC += heavyC
			blocksD += heavyD
			blocksE += heavyE
			stalesD += compD
			stalesE += compE
			stalesF += compF
			heavyA = 0
			heavyB = 0
			heavyC = 0
			heavyD = 0
			heavyE = 0
			compD = 0
			compE = 0
			compF = 0
			compDepth = 0
			heavyDepth = 0
		}
		A = 0
		B = 0
		C = 0
		D = 0
		E = 0
		F = 0
	}
	fmt.Println(blocksA, blocksB, blocksC, blocksD, blocksE, blocksF)
	fmt.Println(stalesA, stalesB, stalesC, stalesD, stalesE, stalesF)
	// Blocks Results: 12167, 12504, 12801, 4949, 4958, 2328 
	// Stales Results:    56,     0,     0,   66,   57,   64
	//
	// Conclusion:
	//
	// 49,950 total blocks were produced in the simulation. Of those, 243 were
	// stale, or about 0.4%.
	//
	// Because of the stale blocks, miners 'B' and 'C' will see a 0.4% increase
	// in total revenue. Miner 'A' sees almost no change in total revenue - the
	// 56 stale blocks are completely offset by the other stale blocks created
	// on the network.
	//
	// Miners 'D' and 'E' see a 0.8% reduction in revenue. Because mining is
	// such a low margin activity, this corresponds to a more significant
	// reduction in profit. This may be enough to deter investment.
	//
	// Miner 'F' sees a 2.2% reduction in revenue. This is significant,
	// especially given that other economy-of-scale related costs mean that the
	// smaller miner is already at a disadvantage. The effect on the profit
	// margin is greatly magnified as the majority of the revenue goes straight
	// to paying for mining infrastructure (electricty). Small miners without
	// access to substantial backbones are vulnerable to block propagation
	// manipulation attacks.
	//
	// Assuming that 'F' declares non-viability, 'A' will see a 5% increase in
	// overall revenue thanks to the elimination of an opponent in a 0-sum game
	// (all other miners will also see this revenue bump).
	//
	// Ignoring many other issues related to uncapped block sizes, this
	// demonstrates that, in network topologies where the majority of the
	// mining power is substantially better connected than a minority, the
	// malicious actions of a single miner can be sufficient to drive the
	// poorly connected minority hashrate out of business, without negatively
	// impacting the revenue of the malicious actor.
}
