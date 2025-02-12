/*
Copyright © 2021 Zoraiz Hassan <hzoraiz8@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image_conversions

import (
	"fmt"
	"image"
	"image/color"

	"github.com/TheZoraiz/ascii-image-converter/aic_package/winsize"
	"github.com/disintegration/imaging"
)

type AsciiPixel struct {
	charDepth      uint32
	grayscaleValue [3]uint32
	rgbValue       [3]uint32
}

func resizeForBraille(asciiWidth, asciiHeight int) (int, int) {
	return asciiWidth * 2, asciiHeight * 4
}

/*
This function shrinks the passed image according to passed dimensions or terminal
size if none are passed. Stores each pixel's grayscale and RGB values in an AsciiPixel
instance to simplify getting numeric data for ASCII character comparison.

The returned 2D AsciiPixel slice contains each corresponding pixel's values. Grayscale value
ranges from 0 to 65535, while RGB values are separate.
*/
func ConvertToAsciiPixels(img image.Image, dimensions []int, width, height int, flipX, flipY, full, isBraille bool) ([][]AsciiPixel, error) {

	var asciiWidth, asciiHeight int
	var smallImg image.Image

	terminalWidth, terminalHeight, err := winsize.GetTerminalSize()
	if err != nil {
		return nil, err
	}

	if full {
		asciiWidth = terminalWidth - 1

		// Passing 0 in place of width keeps the original image's aspect ratio
		smallImg = imaging.Resize(img, asciiWidth, 0, imaging.Lanczos)
		asciiHeight = smallImg.Bounds().Max.Y - smallImg.Bounds().Min.Y

		// To fix aspect ratio in eventual ascii art
		asciiHeight = int(0.5 * float64(asciiHeight))

		if isBraille {
			asciiWidth, asciiHeight = resizeForBraille(asciiWidth, asciiHeight)
		}
		smallImg = imaging.Resize(img, asciiWidth, asciiHeight, imaging.Lanczos)

	} else if (width != 0 || height != 0) && len(dimensions) == 0 {
		// If either width or height is set and dimensions aren't given

		if width > terminalWidth-1 {
			return nil, fmt.Errorf("set width must be lower than terminal width")
		}

		if width != 0 && height == 0 {
			// If width is set and height is not set, use width to calculate aspect ratio

			asciiWidth = width

			smallImg = imaging.Resize(img, asciiWidth, 0, imaging.Lanczos)
			asciiHeight = smallImg.Bounds().Max.Y - smallImg.Bounds().Min.Y

			asciiHeight = int(0.5 * float64(asciiHeight))
			if asciiHeight == 0 {
				asciiHeight = 1
			}

		} else if height != 0 && width == 0 {
			// If height is set and width is not set, use height to calculate aspect ratio

			asciiHeight = height

			smallImg = imaging.Resize(img, 0, asciiHeight, imaging.Lanczos)
			asciiWidth = smallImg.Bounds().Max.X - smallImg.Bounds().Min.X

			asciiWidth = int(2 * float64(asciiWidth))

			if asciiWidth > terminalWidth-1 {
				return nil, fmt.Errorf("width calculated with aspect ratio exceeds terminal width")
			}

		} else {
			return nil, fmt.Errorf("both width and height can't be set. Use dimensions instead")
		}

		if isBraille {
			asciiWidth, asciiHeight = resizeForBraille(asciiWidth, asciiHeight)
		}
		smallImg = imaging.Resize(img, asciiWidth, asciiHeight, imaging.Lanczos)

	} else if len(dimensions) == 0 {
		// This condition calculates aspect ratio according to terminal height

		asciiHeight = terminalHeight - 1

		smallImg = imaging.Resize(img, 0, asciiHeight, imaging.Lanczos)
		asciiWidth = smallImg.Bounds().Max.X - smallImg.Bounds().Min.X

		// To fix aspect ratio in eventual ascii art
		asciiWidth = int(2 * float64(asciiWidth))

		// If ascii width exceeds terminal width, change ratio with respect to terminal width
		if asciiWidth >= terminalWidth {
			asciiWidth = terminalWidth - 1

			smallImg = imaging.Resize(img, asciiWidth, 0, imaging.Lanczos)

			asciiHeight = smallImg.Bounds().Max.Y - smallImg.Bounds().Min.Y

			// To fix aspect ratio in eventual ascii art
			asciiHeight = int(0.5 * float64(asciiHeight))
		}

		if isBraille {
			asciiWidth, asciiHeight = resizeForBraille(asciiWidth, asciiHeight)
		}
		smallImg = imaging.Resize(img, asciiWidth, asciiHeight, imaging.Lanczos)

	} else {
		asciiWidth = dimensions[0]
		asciiHeight = dimensions[1]

		if isBraille {
			asciiWidth, asciiHeight = resizeForBraille(asciiWidth, asciiHeight)
		}
		smallImg = imaging.Resize(img, asciiWidth, asciiHeight, imaging.Lanczos)
	}

	// Repeated despite being in cmd/root.go to maintain support for library
	//
	// If there are passed dimensions, check whether the width exceeds terminal width
	if len(dimensions) > 0 && !full {
		if dimensions[0] > terminalWidth-1 {
			return nil, fmt.Errorf("set width must be lower than terminal width")
		}
	}

	var imgSet [][]AsciiPixel

	b := smallImg.Bounds()

	// These nested loops iterate through each pixel of resized image and get an AsciiPixel instance
	for y := b.Min.Y; y < b.Max.Y; y++ {

		var temp []AsciiPixel
		for x := b.Min.X; x < b.Max.X; x++ {

			oldPixel := smallImg.At(x, y)
			grayPixel := color.GrayModel.Convert(oldPixel)

			r1, g1, b1, _ := grayPixel.RGBA()
			charDepth := r1 / 257 // Only Red is needed from RGB for charDepth in AsciiPixel since they have the same value for grayscale images
			r1 = uint32(r1 / 257)
			g1 = uint32(g1 / 257)
			b1 = uint32(b1 / 257)

			// Get co1ored RGB values of original pixel for rgbValue in AsciiPixel
			r2, g2, b2, _ := oldPixel.RGBA()
			r2 = uint32(r2 / 257)
			g2 = uint32(g2 / 257)
			b2 = uint32(b2 / 257)

			temp = append(temp, AsciiPixel{
				charDepth:      charDepth,
				grayscaleValue: [3]uint32{r1, g1, b1},
				rgbValue:       [3]uint32{r2, g2, b2},
			})

		}
		imgSet = append(imgSet, temp)
	}

	// This rarely affects performance since the ascii art 2D slice size isn't that large
	if flipX || flipY {
		imgSet = reverse(imgSet, flipX, flipY)
	}

	return imgSet, nil
}

func reverse(imgSet [][]AsciiPixel, flipX, flipY bool) [][]AsciiPixel {

	if flipX {
		for _, row := range imgSet {
			for i, j := 0, len(row)-1; i < j; i, j = i+1, j-1 {
				row[i], row[j] = row[j], row[i]
			}
		}
	}

	if flipY {
		for i, j := 0, len(imgSet)-1; i < j; i, j = i+1, j-1 {
			imgSet[i], imgSet[j] = imgSet[j], imgSet[i]
		}
	}

	return imgSet
}
