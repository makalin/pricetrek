package utils

import (
	"fmt"
	"strings"
)

// Sparkline characters for different price ranges
const (
	sparklineChars = "▁▂▃▄▅▆▇█"
)

// GenerateSparkline creates a sparkline visualization from price data
func GenerateSparkline(prices []float64, width int) string {
	if len(prices) == 0 {
		return ""
	}

	if width <= 0 {
		width = len(prices)
	}

	// Normalize prices to 0-7 range for sparkline characters
	min, max := findMinMax(prices)
	if min == max {
		return strings.Repeat("▁", width)
	}

	sparkline := make([]rune, 0, width)
	step := float64(len(prices)) / float64(width)

	for i := 0; i < width; i++ {
		index := int(float64(i) * step)
		if index >= len(prices) {
			index = len(prices) - 1
		}

		normalized := (prices[index] - min) / (max - min)
		charIndex := int(normalized * float64(len(sparklineChars) - 1))
		if charIndex < 0 {
			charIndex = 0
		}
		if charIndex >= len(sparklineChars) {
			charIndex = len(sparklineChars) - 1
		}

		sparkline = append(sparkline, rune(sparklineChars[charIndex]))
	}

	return string(sparkline)
}

// GeneratePriceChart creates a more detailed price chart
func GeneratePriceChart(prices []float64, width, height int) string {
	if len(prices) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	min, max := findMinMax(prices)
	if min == max {
		return strings.Repeat("▁", width)
	}

	// Create a 2D grid
	grid := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}

	step := float64(len(prices)) / float64(width)
	for i := 0; i < width; i++ {
		index := int(float64(i) * step)
		if index >= len(prices) {
			index = len(prices) - 1
		}

		normalized := (prices[index] - min) / (max - min)
		row := int((1.0 - normalized) * float64(height-1))
		if row < 0 {
			row = 0
		}
		if row >= height {
			row = height - 1
		}

		grid[row][i] = true
	}

	// Convert grid to string
	var result strings.Builder
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if grid[i][j] {
				result.WriteString("█")
			} else {
				result.WriteString(" ")
			}
		}
		if i < height-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// FormatPrice formats a price with currency symbol
func FormatPrice(price float64, currency string) string {
	switch currency {
	case "USD":
		return fmt.Sprintf("$%.2f", price)
	case "EUR":
		return fmt.Sprintf("€%.2f", price)
	case "GBP":
		return fmt.Sprintf("£%.2f", price)
	case "TRY":
		return fmt.Sprintf("₺%.2f", price)
	case "JPY":
		return fmt.Sprintf("¥%.0f", price)
	case "INR":
		return fmt.Sprintf("₹%.2f", price)
	default:
		return fmt.Sprintf("%.2f %s", price, currency)
	}
}

// CalculatePriceChange calculates percentage change between two prices
func CalculatePriceChange(oldPrice, newPrice float64) float64 {
	if oldPrice == 0 {
		return 0
	}
	return ((newPrice - oldPrice) / oldPrice) * 100
}

// CalculateMovingAverage calculates moving average for a slice of prices
func CalculateMovingAverage(prices []float64, window int) []float64 {
	if len(prices) < window || window <= 0 {
		return nil
	}

	result := make([]float64, len(prices)-window+1)
	for i := 0; i < len(result); i++ {
		sum := 0.0
		for j := 0; j < window; j++ {
			sum += prices[i+j]
		}
		result[i] = sum / float64(window)
	}

	return result
}

// findMinMax finds minimum and maximum values in a slice
func findMinMax(prices []float64) (min, max float64) {
	if len(prices) == 0 {
		return 0, 0
	}

	min, max = prices[0], prices[0]
	for _, price := range prices[1:] {
		if price < min {
			min = price
		}
		if price > max {
			max = price
		}
	}
	return min, max
}

// CalculateStats calculates basic statistics for price data
func CalculateStats(prices []float64) (min, max, avg, median float64) {
	if len(prices) == 0 {
		return 0, 0, 0, 0
	}

	min, max = findMinMax(prices)
	
	// Calculate average
	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	avg = sum / float64(len(prices))

	// Calculate median
	sorted := make([]float64, len(prices))
	copy(sorted, prices)
	
	// Simple bubble sort for median calculation
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	} else {
		median = sorted[len(sorted)/2]
	}

	return min, max, avg, median
}