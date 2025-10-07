package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	fec "fec-analysis"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	fmt.Println("FEC Loss Models Probability Printer")
	fmt.Println("===================================")
	fmt.Println()

	// Create output directory if it doesn't exist
	outputDir := "loss-models"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	ge := fec.NewGilbertElliotLossModel(0.05, 0.7, 0.05, 0.2)
	// Define loss models to compare
	lossModels := []struct {
		name  string
		model fec.LossModel
	}{
		{"RAND_10", fec.NewRandomLossModel(ge.GetAverageLossProbability())},
		{"G-E_2", ge},
	}

	// Generate loss model comparison for masks of different lengths
	fmt.Printf("Analyzing loss models for different mask lengths...\n")

	filename := filepath.Join(outputDir, "loss_models_analysis.txt")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}

	// Write header to file
	fmt.Fprintf(file, "Loss Model Analysis for Masks of Different Lengths\n")
	fmt.Fprintf(file, "=================================================\n")
	fmt.Fprintf(file, "\n")

	// Print average loss rates for each model
	fmt.Fprintf(file, "Loss Model Average Loss Rates:\n")
	fmt.Fprintf(file, "------------------------------\n")
	for _, lm := range lossModels {
		avgLoss := lm.model.GetAverageLossProbability()
		fmt.Fprintf(file, "%-10s: %.8f (%.4f%%)\n", lm.name, avgLoss, avgLoss*100)
	}
	fmt.Fprintf(file, "\n")

	masksAnalyzed := 0

	for N := 1; N <= 12; N++ {
		// Analyze with different loss models
		printLossModelAnalysis(file, N, lossModels)
		masksAnalyzed += (1 << N) // 2^N masks for length N
	}

	file.Close()
	fmt.Printf("Analyzed %d masks in %s\n", masksAnalyzed, filename)

	// Generate probability density analysis
	fmt.Printf("Generating probability density analysis...\n")
	generateProbabilityDensityAnalysis(outputDir, lossModels)

	fmt.Println("\nLoss model analysis complete!")
}

// printLossModelAnalysis analyzes loss models for given mask length N
func printLossModelAnalysis(file *os.File, N int, lossModels []struct {
	name  string
	model fec.LossModel
}) {
	fmt.Fprintf(file, "Mask Length N=%d Analysis\n", N)
	fmt.Fprintf(file, "%s\n", repeatChar('-', 50))
	fmt.Fprintf(file, "Total packets: %d\n", N)
	fmt.Fprintf(file, "\n")

	// Print scenario probabilities for all delivery patterns
	fmt.Fprintf(file, "Mask Probabilities:\n")
	fmt.Fprintf(file, "%-15s", "Pattern")
	for _, lm := range lossModels {
		fmt.Fprintf(file, " %12s", lm.name)
	}
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "%s\n", repeatChar('-', 15+13*len(lossModels)))

	// Show probabilities for all possible masks
	totalMasks := 1 << N

	// Track probability sums for verification
	probabilitySums := make([]float64, len(lossModels))
	lossPacketCounts := make([]float64, len(lossModels))

	for mask := 0; mask < totalMasks; mask++ {
		maskDesc := formatMask(mask, N)
		fmt.Fprintf(file, "%-15s", maskDesc)

		lostPackets := countLostPackets(mask, N)

		for i, lm := range lossModels {
			prob := lm.model.CalculateProbability(mask, N)
			fmt.Fprintf(file, " %12.8f", prob)

			// Accumulate for verification
			probabilitySums[i] += prob
			lossPacketCounts[i] += prob * float64(lostPackets)
		}
		fmt.Fprintf(file, "\n")
	}

	// Print probability sum verification
	fmt.Fprintf(file, "\nProbability Sum Verification:\n")
	fmt.Fprintf(file, "%s\n", repeatChar('-', 40))
	for i, lm := range lossModels {
		fmt.Fprintf(file, "%-10s: Sum=%.10f (Error: %.2e)\n",
			lm.name, probabilitySums[i], probabilitySums[i]-1.0)
	}

	// Print experimental loss probability calculation
	fmt.Fprintf(file, "\nExperimental Loss Probability:\n")
	fmt.Fprintf(file, "%s\n", repeatChar('-', 45))
	fmt.Fprintf(file, "%-10s %15s %15s %15s\n", "Model", "Theoretical", "Experimental", "Error")
	fmt.Fprintf(file, "%s\n", repeatChar('-', 65))

	for i, lm := range lossModels {
		theoretical := lm.model.GetAverageLossProbability()
		experimental := lossPacketCounts[i] / float64(N)
		error := experimental - theoretical

		fmt.Fprintf(file, "%-10s %15.10f %15.10f %15.2e\n",
			lm.name, theoretical, experimental, error)
	}

	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "%s\n\n", repeatChar('=', 80))
}

// generateProbabilityDensityAnalysis analyzes probability density by number of lost packets
func generateProbabilityDensityAnalysis(outputDir string, lossModels []struct {
	name  string
	model fec.LossModel
}) {
	// Analyze for different packet lengths
	maxN := 10 // Analyze up to N=10 packets

	// Create probability density file
	densityFile := filepath.Join(outputDir, "probability_density_analysis.txt")
	file, err := os.Create(densityFile)
	if err != nil {
		fmt.Printf("Error creating density file %s: %v\n", densityFile, err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "Probability Density Analysis by Number of Lost Packets\n")
	fmt.Fprintf(file, "=====================================================\n\n")

	// Store data for plotting
	plotData := make(map[string]map[int][]float64) // model -> N -> probabilities by lost count

	for _, lm := range lossModels {
		fmt.Fprintf(file, "Loss Model: %s (Avg Loss: %.6f)\n", lm.name, lm.model.GetAverageLossProbability())
		fmt.Fprintf(file, "%s\n", repeatChar('=', 50))

		plotData[lm.name] = make(map[int][]float64)

		for N := 2; N <= maxN; N++ {
			fmt.Fprintf(file, "\nPacket Length N=%d:\n", N)
			fmt.Fprintf(file, "Lost Packets\tProbability\tCumulative\n")
			fmt.Fprintf(file, "%s\n", repeatChar('-', 40))

			// Calculate probability for each number of lost packets
			lostPacketProbs := make([]float64, N+1) // 0 to N lost packets

			totalMasks := 1 << N
			for mask := 0; mask < totalMasks; mask++ {
				lostCount := countLostPackets(mask, N)
				prob := lm.model.CalculateProbability(mask, N)
				lostPacketProbs[lostCount] += prob
			}

			// Print and accumulate for plotting
			cumulative := 0.0
			for lostCount := 0; lostCount <= N; lostCount++ {
				cumulative += lostPacketProbs[lostCount]
				fmt.Fprintf(file, "%d\t\t%.8f\t%.8f\n", lostCount, lostPacketProbs[lostCount], cumulative)
			}

			// Store for plotting
			plotData[lm.name][N] = lostPacketProbs
		}
		fmt.Fprintf(file, "\n")
	}

	fmt.Printf("Probability density analysis saved to: %s\n", densityFile)

	// Create plots
	createProbabilityDensityPlots(outputDir, plotData, lossModels)
}

// createProbabilityDensityPlots creates plots for probability density distributions
func createProbabilityDensityPlots(outputDir string, plotData map[string]map[int][]float64, lossModels []struct {
	name  string
	model fec.LossModel
}) {
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 0, G: 0, B: 255, A: 255},   // Blue
		{R: 0, G: 128, B: 0, A: 255},   // Green
		{R: 128, G: 0, B: 128, A: 255}, // Purple
	}

	// Create plot only for N=10 with regular scale
	N := 10
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Probability Density by Lost Packets (N=%d)", N)
	p.X.Label.Text = "Number of Lost Packets"
	p.Y.Label.Text = "Probability"

	// Create side-by-side bar charts
	barWidth := vg.Points(20) // Wider bars since we only have one plot

	for i, lm := range lossModels {
		if data, exists := plotData[lm.name][N]; exists {
			// Position bars side by side instead of overlapping
			offsetDirection := float64(i) - 0.5*(float64(len(lossModels))-1)

			// Create values for bar chart
			var values plotter.Values
			for _, prob := range data {
				values = append(values, prob)
			}

			if len(values) > 0 {
				// Create histogram bars
				bars, err := plotter.NewBarChart(values, barWidth)
				if err != nil {
					fmt.Printf("Error creating bars for %s: %v\n", lm.name, err)
					continue
				}

				bars.Offset = vg.Points(offsetDirection * 25) // Wider spacing for better visibility
				bars.Color = colors[i%len(colors)]
				bars.LineStyle.Width = vg.Points(1)

				p.Add(bars)
				p.Legend.Add(lm.name, bars)
			}
		}
	}

	// Save plot
	filename := filepath.Join(outputDir, "density_plot_N10.png")
	if err := p.Save(10*vg.Inch, 8*vg.Inch, filename); err != nil {
		fmt.Printf("Error saving plot %s: %v\n", filename, err)
	} else {
		fmt.Printf("Density plot saved: %s\n", filename)
	}

	// Skip combined plot - only generating N=10 log plot
}

// createCombinedDensityPlot creates a combined plot showing multiple N values
func createCombinedDensityPlot(outputDir string, plotData map[string]map[int][]float64, lossModels []struct {
	name  string
	model fec.LossModel
}, colors []color.RGBA) {
	p := plot.New()
	p.Title.Text = "Probability Density Comparison Across Different Packet Lengths"
	p.X.Label.Text = "Number of Lost Packets"
	p.Y.Label.Text = "Probability"

	// Plot for N=5 as a representative case
	N := 5
	barWidth := vg.Points(15) // Narrower bars to fit side by side

	for i, lm := range lossModels {
		if data, exists := plotData[lm.name][N]; exists {
			// Create values for bar chart
			var values plotter.Values
			for _, prob := range data {
				values = append(values, prob)
			}

			// Create histogram bars
			bars, err := plotter.NewBarChart(values, barWidth)
			if err == nil {
				// Position bars side by side instead of overlapping
				offsetDirection := float64(i) - 0.5*(float64(len(lossModels))-1)
				bars.Offset = vg.Points(offsetDirection * 18) // 18 points spacing between bars

				bars.Color = colors[i%len(colors)]
				bars.LineStyle.Width = vg.Points(1)

				p.Add(bars)
				p.Legend.Add(fmt.Sprintf("%s (N=%d)", lm.name, N), bars)
			}
		}
	}

	// Save combined plot
	filename := filepath.Join(outputDir, "density_plot_combined.png")
	if err := p.Save(10*vg.Inch, 7*vg.Inch, filename); err != nil {
		fmt.Printf("Error saving combined plot %s: %v\n", filename, err)
	} else {
		fmt.Printf("Combined density plot saved: %s\n", filename)
	}
}

// countLostPackets counts the number of lost packets in a mask
func countLostPackets(mask, N int) int {
	lostCount := 0
	for i := 0; i < N; i++ {
		if (mask & (1 << i)) == 0 {
			lostCount++
		}
	}
	return lostCount
}

// formatMask formats a mask as a readable bit pattern
func formatMask(mask, N int) string {
	// Create bit pattern description
	desc := ""

	for i := 0; i < N; i++ {
		if (mask & (1 << i)) != 0 {
			desc += "1"
		} else {
			desc += "0"
		}
	}

	return desc
}

// truncateName truncates a name to fit in the specified width
func truncateName(name string, width int) string {
	if len(name) <= width {
		return name
	}
	return name[:width-3] + "..."
}

// repeatChar repeats a character n times
func repeatChar(char rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = char
	}
	return string(result)
}
