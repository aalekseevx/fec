package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	fec "fec-analysis"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type ConfigResult struct {
	N            int
	K            int
	Overhead     float64
	Scenarios    int
	RecoveryProb float64
	LossProb     float64
}

func main() {
	fmt.Println("FEC Recovery Graph Analysis")
	fmt.Println("===========================")
	fmt.Println()

	// Generate test configurations (N, K pairs) - all existing masks up to N=12
	var configs []struct {
		N int
		K int
	}
	for N := 1; N <= 12; N++ {
		for K := 1; K <= N; K++ {
			configs = append(configs, struct {
				N int
				K int
			}{N, K})
		}
	}

	// Test all mask types
	maskTypes := []struct {
		name    string
		factory fec.MaskFactory
	}{
		{"Bursty", &fec.GoogleBurstyMaskFactory{}},
		{"Random", &fec.GoogleRandomMaskFactory{}},
		{"Interleaved", &fec.InterleavedMaskFactory{}},
	}

	// Test different loss probabilities
	lossProbabilities := []float64{0.01, 0.05, 0.1, 0.2, 0.3}

	// Collect all results for plotting
	allResults := make(map[string][]ConfigResult)

	for _, maskType := range maskTypes {
		fmt.Printf("%s Masks:\n", maskType.name)
		fmt.Println("Loss P\tOverhead\tN\tK\tRecovery Prob")
		fmt.Println("─────────────────────────────────────────────")

		var results []ConfigResult

		// Pre-calculate BFS results for each configuration
		type ConfigBFS struct {
			config       struct{ N, K int }
			reachable    []int
			totalPackets int
			overhead     float64
			scenarios    int
		}

		var configBFSResults []ConfigBFS

		for _, config := range configs {
			// Create mask
			mask, err := maskType.factory.CreateMask(config.N, config.K)
			if err != nil {
				continue // Skip unsupported configurations
			}

			// Create recovery graph
			graph := fec.NewRecoveryGraph(mask)
			totalPackets := config.N + config.K

			// Generate "good" vertices: first N bits are 1, next K bits are any, rest are 0
			var goodVertices []int
			allMediaPackets := (1 << config.N) - 1 // First N bits set to 1

			// Generate all combinations of FEC packet delivery states
			for fecState := 0; fecState < (1 << config.K); fecState++ {
				goodVertex := allMediaPackets | (fecState << config.N)
				goodVertices = append(goodVertices, goodVertex)
			}

			// Run multi-source BFS from all good vertices (once per configuration)
			reachable := fec.BFS(graph, goodVertices)
			overhead := float64(config.K) * 100.0 / float64(config.N)

			configBFSResults = append(configBFSResults, ConfigBFS{
				config:       config,
				reachable:    reachable,
				totalPackets: totalPackets,
				overhead:     overhead,
				scenarios:    graph.NumVertices(),
			})
		}

		// Now calculate recovery probabilities for each loss model using pre-computed BFS results
		for _, lossP := range lossProbabilities {
			lossModel := fec.NewRandomLossModel(lossP)

			for _, configBFS := range configBFSResults {
				// Calculate recovery probability by summing probabilities of recovered scenarios
				recoveryProb := 0.0
				for _, vertex := range configBFS.reachable {
					prob := lossModel.CalculateProbability(vertex, configBFS.totalPackets)
					recoveryProb += prob
				}

				// Normalize by taking the Nth root to account for needing all N media packets
				if recoveryProb > 0 && configBFS.config.N > 0 {
					recoveryProb = math.Pow(recoveryProb, 1.0/float64(configBFS.config.N))
				}

				results = append(results, ConfigResult{
					N:            configBFS.config.N,
					K:            configBFS.config.K,
					Overhead:     configBFS.overhead,
					Scenarios:    configBFS.scenarios,
					RecoveryProb: recoveryProb,
					LossProb:     lossP,
				})
			}
		}

		// Sort by (loss probability, overhead, N)
		sort.Slice(results, func(i, j int) bool {
			if results[i].LossProb != results[j].LossProb {
				return results[i].LossProb < results[j].LossProb
			}
			if results[i].Overhead != results[j].Overhead {
				return results[i].Overhead < results[j].Overhead
			}
			return results[i].N < results[j].N
		})

		// Print sorted results
		for _, result := range results {
			fmt.Printf("%.2f\t%.1f%%\t\t%d\t%d\t%.6f\n",
				result.LossProb, result.Overhead, result.N, result.K, result.RecoveryProb)
		}
		fmt.Println()

		// Store results for plotting
		allResults[maskType.name] = results
	}

	// Create plots
	createPlots(allResults, lossProbabilities)
}

func createPlots(allResults map[string][]ConfigResult, lossProbabilities []float64) {
	// Create a plot for each loss probability
	for _, lossP := range lossProbabilities {
		p := plot.New()
		p.Title.Text = fmt.Sprintf("Recovery Probability vs Overhead (Loss P = %.2f)", lossP)
		p.X.Label.Text = "Overhead (%)"
		p.Y.Label.Text = "Recovery Probability"

		colors := []color.RGBA{
			{R: 255, G: 0, B: 0, A: 255},     // Red for Bursty
			{R: 0, G: 0, B: 255, A: 255},     // Blue for Random
			{R: 0, G: 128, B: 0, A: 255},     // Green for Interleaved
			{R: 128, G: 128, B: 128, A: 255}, // Gray for No FEC
		}

		colorIndex := 0
		for maskName, results := range allResults {
			// Filter results for this loss probability and preprocess to keep only highest recovery for each overhead
			overheadMap := make(map[float64]float64) // overhead -> max recovery probability
			for _, result := range results {
				if result.LossProb == lossP {
					if existing, exists := overheadMap[result.Overhead]; !exists || result.RecoveryProb > existing {
						overheadMap[result.Overhead] = result.RecoveryProb
					}
				}
			}

			// Convert map to sorted points
			var points plotter.XYs
			for overhead, recoveryProb := range overheadMap {
				points = append(points, plotter.XY{
					X: overhead,
					Y: recoveryProb,
				})
			}

			// Sort points by overhead for proper line plotting
			sort.Slice(points, func(i, j int) bool {
				return points[i].X < points[j].X
			})

			// Post-process to ensure monotonically increasing recovery probability
			// Skip points where recovery probability decreases as overhead increases
			var monotonicPoints plotter.XYs
			if len(points) > 0 {
				monotonicPoints = append(monotonicPoints, points[0]) // Always include first point

				for i := 1; i < len(points); i++ {
					// Only include point if recovery probability is >= previous point
					if points[i].Y >= monotonicPoints[len(monotonicPoints)-1].Y {
						monotonicPoints = append(monotonicPoints, points[i])
					}
					// Skip points that would decrease recovery probability
				}
			}

			// Use monotonic points for plotting
			points = monotonicPoints

			if len(points) > 0 {
				// Create line plot
				line, err := plotter.NewLine(points)
				if err != nil {
					fmt.Printf("Error creating line for %s: %v\n", maskName, err)
					continue
				}
				line.Color = colors[colorIndex%len(colors)]
				line.Width = vg.Points(2)

				// Create scatter plot
				scatter, err := plotter.NewScatter(points)
				if err != nil {
					fmt.Printf("Error creating scatter for %s: %v\n", maskName, err)
					continue
				}
				scatter.Color = colors[colorIndex%len(colors)]
				scatter.Radius = vg.Points(3)

				p.Add(line, scatter)
				p.Legend.Add(maskName, line, scatter)
			}
			colorIndex++
		}

		// Save the plot
		filename := fmt.Sprintf("img/recovery_plot_p%.2f.png", lossP)
		if err := p.Save(8*vg.Inch, 6*vg.Inch, filename); err != nil {
			fmt.Printf("Error saving plot %s: %v\n", filename, err)
		} else {
			fmt.Printf("Plot saved: %s\n", filename)
		}
	}
}
