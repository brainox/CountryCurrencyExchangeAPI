package helpers

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"country-currency-exchange-api/models"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// generateSummaryImage creates a summary image with country statistics
func GenerateSummaryImage(countries []models.Country, lastRefreshed string) error {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll("cache", 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %v", err)
	}

	// Sort countries by GDP in descending order
	sortedCountries := make([]models.Country, len(countries))
	copy(sortedCountries, countries)

	sort.Slice(sortedCountries, func(i, j int) bool {
		return sortedCountries[i].EstimatedGDP > sortedCountries[j].EstimatedGDP
	})

	// Get top 5 countries
	top5 := sortedCountries
	if len(sortedCountries) > 5 {
		top5 = sortedCountries[:5]
	}

	// Create image
	const width = 1400
	const height = 800
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background color - dark blue-gray
	bgColor := color.RGBA{30, 41, 59, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw border
	borderColor := color.RGBA{51, 65, 85, 255}
	drawRect(img, 40, 40, width-40, height-40, borderColor, 2)

	// Text colors
	white := color.RGBA{255, 255, 255, 255}
	lightGray := color.RGBA{203, 213, 225, 255}
	green := color.RGBA{74, 222, 128, 255}
	gray := color.RGBA{148, 163, 184, 255}

	// Draw text
	face := basicfont.Face7x13

	// Title
	drawText(img, "Countries API Summary", width/2-150, 120, white, face)

	// Total countries
	totalText := fmt.Sprintf("Total Countries in DB: %d", len(countries))
	drawText(img, totalText, width/2-120, 200, lightGray, face)

	// Top 5 header
	drawText(img, "Top 5 Countries by Estimated GDP (USD):", width/2-220, 280, white, face)

	// List top 5 countries
	yPosition := 350
	for i, country := range top5 {
		// Country name
		countryText := fmt.Sprintf("%d. %s", i+1, country.Name)
		drawText(img, countryText, 150, yPosition, lightGray, face)

		// GDP value (right-aligned)
		gdpText := fmt.Sprintf("$%.2f", country.EstimatedGDP)
		drawText(img, gdpText, width-400, yPosition, green, face)

		yPosition += 60
	}

	// Last refreshed timestamp
	refreshText := fmt.Sprintf("Last Refreshed: %s", lastRefreshed)
	drawText(img, refreshText, width/2-200, height-80, gray, face)

	// Save image
	imagePath := filepath.Join("cache", "summary.png")
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("error creating image file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("error encoding image: %v", err)
	}

	return nil
}

// drawText draws text at the specified position
func drawText(img *image.RGBA, text string, x, y int, col color.Color, face font.Face) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)
}

// drawRect draws a rectangle outline
func drawRect(img *image.RGBA, x1, y1, x2, y2 int, col color.Color, thickness int) {
	for t := 0; t < thickness; t++ {
		// Top
		for x := x1; x <= x2; x++ {
			img.Set(x, y1+t, col)
		}
		// Bottom
		for x := x1; x <= x2; x++ {
			img.Set(x, y2-t, col)
		}
		// Left
		for y := y1; y <= y2; y++ {
			img.Set(x1+t, y, col)
		}
		// Right
		for y := y1; y <= y2; y++ {
			img.Set(x2-t, y, col)
		}
	}
}

// GetSummaryImagePath returns the path to the summary image
func GetSummaryImagePath() string {
	return filepath.Join("cache", "summary.png")
}
