// Lissajous is a small program for rendering Lissajous curves.
//
// See wikipedia for more info:
// https://en.wikipedia.org/wiki/Lissajous_curve
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"math"
	"os"
)

// Imgur color palette. Ranges from imgur grey to white in a smooth 16-color
// gradient.
var palette = []color.Color{
	color.RGBA{0x2c, 0x2f, 0x34, 0xff},
	color.RGBA{0x3b, 0x3e, 0x42, 0xff},
	color.RGBA{0x4a, 0x4d, 0x50, 0xff},
	color.RGBA{0x59, 0x5c, 0x5e, 0xff},
	color.RGBA{0x68, 0x6b, 0x6c, 0xff},
	color.RGBA{0x77, 0x7a, 0x7a, 0xff},
	color.RGBA{0x86, 0x89, 0x88, 0xff},
	color.RGBA{0x95, 0x98, 0x96, 0xff},
	color.RGBA{0xa4, 0xa7, 0x94, 0xff},
	color.RGBA{0xb3, 0xb6, 0xa2, 0xff},
	color.RGBA{0xc2, 0xc5, 0xb0, 0xff},
	color.RGBA{0xd1, 0xd4, 0xbe, 0xff},
	color.RGBA{0xe0, 0xe3, 0xcc, 0xff},
	color.RGBA{0xff, 0xf2, 0xda, 0xff},
	color.RGBA{0xff, 0xf1, 0xe8, 0xff},
	color.RGBA{0xff, 0xff, 0xff, 0xff},
}

// Configuration flags
var (
	outFile   string
	nframes   int
	size      int
	delay     int
	cycles    float64
	xfreq     float64
	xfreqInc  float64
	yfreq     float64
	yfreqInc  float64
	xphase    float64
	xphaseInc float64
	yphase    float64
	yphaseInc float64
	res       float64
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Paint is an rough anti-aliasing rounding paint. Scales by "size", and then
// offsets in all four directions by 0.55 before rounding.
func paint(x, y float64, scale int, canvas [][]int, maxVal int) {
	x = x * float64(scale)
	y = y * float64(scale)
	for _, xoffset := range []float64{-0.55, 0, 0.55} {
		for _, yoffset := range []float64{-0.55, 0, 0.55} {
			intx := scale + int(x+xoffset)
			inty := scale + int(y+yoffset)
			if intx >= len(canvas) || intx < 0 {
				continue
			}
			if inty >= len(canvas) || inty < 0 {
				continue
			}
			canvas[intx][inty] = min(canvas[intx][inty]+1, maxVal)
		}
	}
}

// lissajous computes the lissajous curve, and plots it onto several gif frames.
func lissajous(out io.Writer) error {
	anim := gif.GIF{LoopCount: nframes}
	for i := 0; i < nframes; i++ {
		rect := image.Rect(0, 0, 2*size+1, 2*size-1)
		img := image.NewPaletted(rect, palette)

		// Initialize the pixel array.
		lisa := make([][]int, 2*size)
		for i := range lisa {
			lisa[i] = make([]int, 2*size)
		}

		// Compute the values at each pixel.
		for t := 0.0; t < cycles*2*math.Pi; t += res {
			x := math.Sin(t*xfreq + xphase)
			y := math.Sin(t*yfreq + yphase)
			paint(x, y, size-2, lisa, 15)
		}

		// Render it as a gif frame.
		for x := range lisa {
			for y := range lisa[x] {
				img.Set(x, y, palette[lisa[x][y]])
			}
		}

		// Increment for the next frame.
		xphase += xphaseInc
		yphase += yphaseInc
		xfreq += xfreqInc
		yfreq += yfreqInc

		// Store the neccesary data.
		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, img)
		fmt.Printf("\rRendered frame %d of %d", i, nframes)
	}
	fmt.Printf("\rRendered frame %d of %d\n", nframes, nframes)
	return gif.EncodeAll(out, &anim)
}

func init() {
	flag.StringVar(&outFile, "outfile", "out.gif", "Name of file to store output gif.")
	flag.IntVar(&nframes, "nframes", 1, "Number of frames to render.")
	flag.IntVar(&size, "size", 100, "Radius of the image.")
	flag.IntVar(&delay, "delay", 8, "Delay between frames (in ms).")
	flag.Float64Var(&cycles, "cycles", 2, "Lenth of the curve's stroke.")
	flag.Float64Var(&xfreq, "xfreq", 5.0, "X frequency.")
	flag.Float64Var(&xfreqInc, "xfreq_inc", 0.00, "X frequency increment per frame.")
	flag.Float64Var(&yfreq, "yfreq", 4.0, "Y frequency.")
	flag.Float64Var(&yfreqInc, "yfreq_inc", 0.00, "Y frequency increment per frame.")
	flag.Float64Var(&xphase, "xphase", 0.0, "X phase.")
	flag.Float64Var(&xphaseInc, "xphase_inc", 0.00, "X phase increment per frame.")
	flag.Float64Var(&yphase, "yphase", 0.0, "Y phase.")
	flag.Float64Var(&yphaseInc, "yphase_inc", 0.01, "Y phase increment per frame.")
	flag.Float64Var(&res, "res", 0.0001, "Angular resolution")

	flag.Parse()
}

func main() {
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("Could not open file %s: %v", outFile, err)
	}
	if err := lissajous(f); err != nil {
		fmt.Printf("Error encoding GIF: %v", err)
	}
}
