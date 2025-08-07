package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	// Canvas dimensions (9:16 aspect ratio)
	canvasWidth  = 1080
	canvasHeight = 1920

	// Safe zones
	topMargin    = 250
	bottomMargin = 250
	sideMargin   = 60

	// Working area
	workingWidth  = canvasWidth - (sideMargin * 2)          // 960px
	workingHeight = canvasHeight - topMargin - bottomMargin // 1420px

	// Poster dimensions (maintaining aspect ratio ~2:3)
	posterWidth  = 180
	posterHeight = 270

	// Grid settings
	postersPerRow  = 3  // 3 posters per row
	posterSpacingX = 10 // Column gap
	posterSpacingY = 10 // Row gap
	borderRadius   = 8
)

type CollageConfig struct {
	Title           string
	Year            int
	BackgroundStart color.RGBA
	BackgroundEnd   color.RGBA
	TextColor       color.RGBA
}

func GenerateCollageImage(history WatchedHistory, config CollageConfig) error {
	// Calculate how many posters fit per frame
	titleAreaHeight := 120
	availableHeight := workingHeight - titleAreaHeight
	maxRowsPerFrame := availableHeight / (posterHeight + posterSpacingY)
	postersPerFrame := maxRowsPerFrame * postersPerRow

	// Calculate number of frames needed
	totalPosters := len(history)
	numFrames := int(math.Ceil(float64(totalPosters) / float64(postersPerFrame)))

	fmt.Printf("Generating %d frame(s) for %d movies (%d posters per frame)\n",
		numFrames, totalPosters, postersPerFrame)

	// Generate each frame
	for frameNum := 0; frameNum < numFrames; frameNum++ {
		startIdx := frameNum * postersPerFrame
		endIdx := startIdx + postersPerFrame
		if endIdx > totalPosters {
			endIdx = totalPosters
		}

		frameMovies := history[startIdx:endIdx]

		err := generateSingleFrame(frameMovies, config, frameNum+1, numFrames)
		if err != nil {
			return fmt.Errorf("error generating frame %d: %v", frameNum+1, err)
		}
	}

	return nil
}

func generateSingleFrame(movies WatchedHistory, config CollageConfig, frameNum, totalFrames int) error {
	// Create canvas
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	// Draw gradient background
	drawGradientBackground(canvas, config.BackgroundStart, config.BackgroundEnd)

	// Draw title and stats
	stats := calculateStats(movies)
	frameTitle := config.Title
	if totalFrames > 1 {
		frameTitle = fmt.Sprintf("%s - Part %d", config.Title, frameNum)
	}
	drawTitle(canvas, frameTitle, stats, config.TextColor)

	// Draw movie posters
	for i, movie := range movies {
		row := i / postersPerRow
		col := i % postersPerRow

		// Calculate position (centering the grid with fixed gaps)
		totalGridWidth := postersPerRow*posterWidth + (postersPerRow-1)*posterSpacingX
		startX := sideMargin + (workingWidth-totalGridWidth)/2

		x := startX + col*(posterWidth+posterSpacingX)
		y := topMargin + 120 + row*(posterHeight+posterSpacingY) // 120px for title area

		// Check if poster would overflow into bottom safe zone
		if y+posterHeight > canvasHeight-bottomMargin {
			log.Printf("Warning: Poster at position (%d,%d) would overflow into safe zone", x, y)
			continue
		}

		// Load and draw poster
		err := drawMoviePoster(canvas, movie, x, y)
		if err != nil {
			log.Printf("Error drawing poster for %s: %v", movie.Movie.Title, err)
			// Draw placeholder
			drawPlaceholderPoster(canvas, movie.Movie.Title, x, y)
		}
	}

	// Save image
	var outputPath string
	if totalFrames > 1 {
		outputPath = fmt.Sprintf("images/movie_collage_%d_part_%d.png", config.Year, frameNum)
	} else {
		outputPath = fmt.Sprintf("images/movie_collage_%d.png", config.Year)
	}

	err := saveImage(canvas, outputPath)
	if err == nil {
		fmt.Printf("✅ Generated frame %d: %s\n", frameNum, outputPath)
	}

	return err
}

func drawGradientBackground(canvas *image.RGBA, startColor, endColor color.RGBA) {
	bounds := canvas.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate gradient ratio (0.0 to 1.0)
		ratio := float64(y-bounds.Min.Y) / float64(bounds.Max.Y-bounds.Min.Y)

		// Interpolate colors
		r := uint8(float64(startColor.R)*(1-ratio) + float64(endColor.R)*ratio)
		g := uint8(float64(startColor.G)*(1-ratio) + float64(endColor.G)*ratio)
		b := uint8(float64(startColor.B)*(1-ratio) + float64(endColor.B)*ratio)
		a := uint8(float64(startColor.A)*(1-ratio) + float64(endColor.A)*ratio)

		gradientColor := color.RGBA{r, g, b, a}

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			canvas.Set(x, y, gradientColor)
		}
	}
}

func drawTitle(canvas *image.RGBA, title string, stats MovieStats, textColor color.RGBA) {
	// Title
	drawText(canvas, title, canvasWidth/2, topMargin/2, 48, textColor, true)

	// Stats - show frame stats if it's a multi-frame collage
	statsText := fmt.Sprintf("%d Movies • %.1f Hours • ⭐ %.1f",
		stats.Count, stats.TotalHours, stats.AvgRating)
	drawText(canvas, statsText, canvasWidth/2, topMargin/2+60, 24, textColor, true)
}

func drawText(canvas *image.RGBA, text string, x, y, size int, textColor color.RGBA, center bool) {
	// This is a simplified text drawing - you might want to use a proper font library
	// For now, using basic font
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	if center {
		// Rough centering calculation
		textWidth := len(text) * (size / 2)
		point.X = fixed.Int26_6((x - textWidth/2) * 64)
	}

	drawer := &font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13, // Using basic font - upgrade this for better results
		Dot:  point,
	}
	drawer.DrawString(text)
}

func drawMoviePoster(canvas *image.RGBA, movie WatchedHistoryElement, x, y int) error {
	// Generate expected filename
	imageFile := movie.WatchedAt.Local().Format("2006-01-02_15-04-05") + ".jpg"
	imagePath := filepath.Join("images", imageFile)

	// Load image
	posterImg, err := loadImage(imagePath)
	if err != nil {
		return err
	}

	// Resize to fit poster dimensions
	resizedPoster := resizeImage(posterImg, posterWidth, posterHeight)

	// Draw poster with border/shadow effect
	drawPosterWithEffects(canvas, resizedPoster, x, y)

	return nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Detect image type and decode
	if strings.HasSuffix(strings.ToLower(path), ".png") {
		return png.Decode(file)
	} else {
		return jpeg.Decode(file)
	}
}

func resizeImage(src image.Image, width, height int) *image.RGBA {
	// Use high-quality Lanczos resampling
	resized := resize.Resize(uint(width), uint(height), src, resize.Lanczos3)

	// Convert to RGBA
	bounds := resized.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, resized, bounds.Min, draw.Src)

	return rgba
}

func drawPosterWithEffects(canvas *image.RGBA, poster *image.RGBA, x, y int) {
	// Create rounded mask for the poster
	roundedPoster := applyBorderRadius(poster, borderRadius)

	// Draw shadow with rounded corners
	shadowOffset := 4
	shadowColor := color.RGBA{0, 0, 0, 100}
	shadowMask := createRoundedRectMask(posterWidth, posterHeight, borderRadius)
	drawShadowWithMask(canvas, shadowMask, x+shadowOffset, y+shadowOffset, shadowColor)

	// Draw border with rounded corners
	borderColor := color.RGBA{255, 255, 255, 200}
	borderThickness := 2
	drawRoundedBorder(canvas, x-borderThickness, y-borderThickness,
		posterWidth+2*borderThickness, posterHeight+2*borderThickness,
		borderRadius+borderThickness, borderColor)

	// Draw rounded poster
	posterRect := image.Rect(x, y, x+posterWidth, y+posterHeight)
	draw.Draw(canvas, posterRect, roundedPoster, image.Point{}, draw.Over)
}

func applyBorderRadius(img *image.RGBA, radius int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isInsideRoundedRect(x, y, bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy(), radius) {
				result.Set(x, y, img.At(x, y))
			} else {
				result.Set(x, y, color.RGBA{0, 0, 0, 0}) // Transparent
			}
		}
	}

	return result
}

func createRoundedRectMask(width, height, radius int) *image.RGBA {
	mask := image.NewRGBA(image.Rect(0, 0, width, height))
	fillColor := color.RGBA{255, 255, 255, 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if isInsideRoundedRect(x, y, 0, 0, width, height, radius) {
				mask.Set(x, y, fillColor)
			}
		}
	}

	return mask
}

func isInsideRoundedRect(x, y, rectX, rectY, width, height, radius int) bool {
	// Check if point is in the main rectangle (excluding corners)
	if x >= rectX+radius && x < rectX+width-radius {
		return y >= rectY && y < rectY+height
	}
	if y >= rectY+radius && y < rectY+height-radius {
		return x >= rectX && x < rectX+width
	}

	// Check corners
	corners := []struct{ cx, cy int }{
		{rectX + radius, rectY + radius},                          // Top-left
		{rectX + width - radius - 1, rectY + radius},              // Top-right
		{rectX + radius, rectY + height - radius - 1},             // Bottom-left
		{rectX + width - radius - 1, rectY + height - radius - 1}, // Bottom-right
	}

	for _, corner := range corners {
		dx := x - corner.cx
		dy := y - corner.cy
		if dx*dx+dy*dy <= radius*radius {
			return true
		}
	}

	return false
}

func drawShadowWithMask(canvas *image.RGBA, mask *image.RGBA, x, y int, shadowColor color.RGBA) {
	maskBounds := mask.Bounds()
	for my := maskBounds.Min.Y; my < maskBounds.Max.Y; my++ {
		for mx := maskBounds.Min.X; mx < maskBounds.Max.X; mx++ {
			if _, _, _, a := mask.At(mx, my).RGBA(); a > 0 {
				canvas.Set(x+mx, y+my, shadowColor)
			}
		}
	}
}

func drawRoundedBorder(canvas *image.RGBA, x, y, width, height, radius int, borderColor color.RGBA) {
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			if isInsideRoundedRect(px, py, 0, 0, width, height, radius) {
				canvas.Set(x+px, y+py, borderColor)
			}
		}
	}
}

func drawPlaceholderPoster(canvas *image.RGBA, title string, x, y int) {
	// Draw rounded placeholder background
	placeholderColor := color.RGBA{200, 200, 200, 255}

	for py := 0; py < posterHeight; py++ {
		for px := 0; px < posterWidth; px++ {
			if isInsideRoundedRect(px, py, 0, 0, posterWidth, posterHeight, borderRadius) {
				canvas.Set(x+px, y+py, placeholderColor)
			}
		}
	}

	// Draw title text (simplified)
	textColor := color.RGBA{100, 100, 100, 255}
	drawText(canvas, title, x+posterWidth/2, y+posterHeight/2, 16, textColor, true)
}

type MovieStats struct {
	Count      int
	TotalHours float64
	AvgRating  float64
}

func calculateStats(history WatchedHistory) MovieStats {
	var totalRating float64
	var ratingCount int
	var totalMinutes int64

	for _, movie := range history {
		if movie.Movie.IDs.Tmdb == 0 {
			continue
		}

		details := GetMovieDetails(int(movie.Movie.IDs.Tmdb))

		if details.VoteAverage > 0 {
			totalRating += details.VoteAverage
			ratingCount++
		}

		if details.Runtime > 0 {
			totalMinutes += details.Runtime
		}
	}

	avgRating := 0.0
	if ratingCount > 0 {
		avgRating = totalRating / float64(ratingCount)
	}

	return MovieStats{
		Count:      len(history),
		TotalHours: float64(totalMinutes) / 60.0,
		AvgRating:  avgRating,
	}
}

func saveImage(img *image.RGBA, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
