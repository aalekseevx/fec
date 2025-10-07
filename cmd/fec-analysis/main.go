package main

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"

	fec "fec-analysis"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type LossModelResult struct {
	Name         string  // "Random" or Gilbert-Elliott variant name
	LossProb     float64 // Average loss probability
	RecoveryProb float64 // Recovery probability for this loss model
}

type ConfigResult struct {
	N                                int
	K                                int
	Overhead                         float64
	Scenarios                        int
	LossModelResults                 []LossModelResult
	MinLostPacketsForNonRecovery     int
	MinConsecutiveLostForNonRecovery int
}

func main() {
	fmt.Println("FEC Recovery Graph Analysis")
	fmt.Println("===========================")
	fmt.Println()

	// Generate test configurations (N, K pairs) - smaller set for testing
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

	// Define single Gilbert-Elliott loss model
	lossModels := []struct {
		name  string
		model fec.LossModel
	}{
		// Gilbert-Elliott model
		{"Gilbert_Elliott", fec.NewGilbertElliotLossModel(0.05, 0.7, 0.05, 0.2)},
	}

	// Collect all results for plotting
	allResults := make(map[string][]ConfigResult)

	for _, maskType := range maskTypes {
		fmt.Printf("%s Masks:\n", maskType.name)

		// Create dynamic header based on available loss models
		header := "Overhead\tN\tK\t"
		for _, lm := range lossModels {
			header += fmt.Sprintf("%s (P=%.2f)\t", lm.name, lm.model.GetAverageLossProbability())
		}
		header += "Min Lost\tMin Consec"
		fmt.Println(header)

		// Create separator line
		separatorLen := len(header) - len(header)/8 // Approximate adjustment for tabs
		separator := ""
		for i := 0; i < separatorLen; i++ {
			separator += "─"
		}
		fmt.Println(separator)

		var results []ConfigResult

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
			scenarios := graph.NumVertices()

			// Calculate recovery characteristics (once per configuration)
			characteristics := fec.CalculateRecoveryCharacteristicsFromReachable(config.N, config.K, reachable)

			// Calculate recovery probabilities for all loss models
			var lossModelResults []LossModelResult
			for _, lossModelConfig := range lossModels {
				// Calculate recovery probability by summing probabilities of recovered scenarios
				recoveryProb := 0.0
				for _, vertex := range reachable {
					prob := lossModelConfig.model.CalculateProbability(vertex, totalPackets)
					recoveryProb += prob
				}

				// Normalize by taking the Nth root to account for needing all N media packets
				if recoveryProb > 0 && config.N > 0 {
					recoveryProb = math.Pow(recoveryProb, 1.0/float64(config.N))
				}

				lossModelResults = append(lossModelResults, LossModelResult{
					Name:         lossModelConfig.name,
					LossProb:     lossModelConfig.model.GetAverageLossProbability(),
					RecoveryProb: recoveryProb,
				})
			}

			// Create single result per configuration with all loss model results
			results = append(results, ConfigResult{
				N:                                config.N,
				K:                                config.K,
				Overhead:                         overhead,
				Scenarios:                        scenarios,
				LossModelResults:                 lossModelResults,
				MinLostPacketsForNonRecovery:     characteristics.MinLostPacketsForNonRecovery,
				MinConsecutiveLostForNonRecovery: characteristics.MinConsecutiveLostForNonRecovery,
			})
		}

		// Sort by (overhead, N) since we now have one row per configuration
		sort.Slice(results, func(i, j int) bool {
			if results[i].Overhead != results[j].Overhead {
				return results[i].Overhead < results[j].Overhead
			}
			return results[i].N < results[j].N
		})

		// Print sorted results
		for _, result := range results {
			// Start with basic config info
			fmt.Printf("%.1f%%\t\t%d\t%d\t", result.Overhead, result.N, result.K)

			// Print recovery probability for each loss model
			for _, lmResult := range result.LossModelResults {
				fmt.Printf("%.6f\t\t", lmResult.RecoveryProb)
			}

			// Print characteristics
			if result.MinLostPacketsForNonRecovery > 0 {
				fmt.Printf("%d\t%d\n", result.MinLostPacketsForNonRecovery, result.MinConsecutiveLostForNonRecovery)
			} else if result.MinLostPacketsForNonRecovery == -1 {
				fmt.Printf("∞\t∞\n")
			} else {
				fmt.Printf("-\t-\n")
			}
		}
		fmt.Println()

		// Store results by mask type for plotting
		allResults[maskType.name] = append(allResults[maskType.name], results...)
	}

	// Create combined plots with both loss models
	createCombinedPlots(allResults)
}

func createCombinedPlots(allResults map[string][]ConfigResult) {
	// Group results by mask type
	resultsByMaskType := make(map[string][]ConfigResult)

	for maskType, results := range allResults {
		resultsByMaskType[maskType] = results
	}

	// Colors for the 3 mask types (bright colors for dark background)
	maskColors := map[string]color.RGBA{
		"Bursty":      {R: 100, G: 200, B: 255, A: 255}, // Cyan/Light Blue
		"Random":      {R: 255, G: 200, B: 100, A: 255}, // Orange/Gold
		"Interleaved": {R: 255, G: 100, B: 150, A: 255}, // Pink/Rose
	}

	// Create single combined plot
	p := plot.New()

	// Dark theme styling with bigger fonts
	textColor := color.RGBA{R: 240, G: 240, B: 240, A: 255} // Light gray text
	p.Title.Text = "Recovery Probability vs Overhead - Gilbert-Elliott Model"
	p.Title.TextStyle.Font.Size = vg.Points(24)
	p.Title.TextStyle.Color = textColor

	p.X.Label.Text = "Overhead (%)"
	p.X.Label.TextStyle.Font.Size = vg.Points(20)
	p.X.Label.TextStyle.Color = textColor
	p.X.Tick.Label.Font.Size = vg.Points(16)
	p.X.Tick.Label.Color = textColor
	p.X.Tick.Color = textColor
	p.X.Color = textColor

	p.Y.Label.Text = "Recovery Probability"
	p.Y.Label.TextStyle.Font.Size = vg.Points(20)
	p.Y.Label.TextStyle.Color = textColor
	p.Y.Tick.Label.Font.Size = vg.Points(16)
	p.Y.Tick.Label.Color = textColor
	p.Y.Tick.Color = textColor
	p.Y.Color = textColor

	// Transparent background
	p.BackgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}

	// Legend styling
	p.Legend.TextStyle.Font.Size = vg.Points(16)
	p.Legend.TextStyle.Color = textColor

	// Process each mask type
	maskTypeOrder := []string{"Bursty", "Random", "Interleaved"}

	for _, maskType := range maskTypeOrder {
		results, exists := resultsByMaskType[maskType]
		if !exists || len(results) == 0 {
			continue
		}

		points := processResultsToPoints(results)
		if len(points) > 0 {
			maskColor := maskColors[maskType]

			// Line
			line, err := plotter.NewLine(points)
			if err == nil {
				line.Color = maskColor
				line.Width = vg.Points(3)

				// Scatter points
				scatter, err := plotter.NewScatter(points)
				if err == nil {
					scatter.Color = maskColor
					scatter.Radius = vg.Points(4)

					p.Add(line, scatter)
					p.Legend.Add(maskType, line, scatter)
				}
			}
		}
	}

	// Save the combined plot with transparent background
	filename := "img/recovery_plot_combined.png"

	// Create canvas with custom transparent background
	width := 12 * vg.Inch
	height := 9 * vg.Inch

	// Create image canvas
	c := vgimg.New(width, height)

	// Make the canvas background transparent
	// Access the underlying image and clear it with transparency
	img := c.Image()
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Draw the plot on the transparent canvas
	dc := draw.New(c)
	p.Draw(dc)

	// Save to file
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer f.Close()

	// Encode as PNG
	if err := png.Encode(f, c.Image()); err != nil {
		fmt.Printf("Error encoding PNG %s: %v\n", filename, err)
	} else {
		fmt.Printf("Combined plot saved: %s\n", filename)
	}
}

func processResultsToPoints(results []ConfigResult) plotter.XYs {
	// Preprocess to keep only highest recovery for each overhead
	overheadMap := make(map[float64]float64) // overhead -> max recovery probability
	for _, result := range results {
		if len(result.LossModelResults) > 0 {
			recoveryProb := result.LossModelResults[0].RecoveryProb
			if existing, exists := overheadMap[result.Overhead]; !exists || recoveryProb > existing {
				overheadMap[result.Overhead] = recoveryProb
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
	var monotonicPoints plotter.XYs
	if len(points) > 0 {
		monotonicPoints = append(monotonicPoints, points[0])

		for i := 1; i < len(points); i++ {
			if points[i].Y >= monotonicPoints[len(monotonicPoints)-1].Y {
				monotonicPoints = append(monotonicPoints, points[i])
			}
		}
	}

	return monotonicPoints
}
