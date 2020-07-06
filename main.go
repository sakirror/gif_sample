package main

import (
	"errors"
	"flag"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/soniakeys/quant/median"
	"golang.org/x/image/draw"
)

type WH struct {
	W int
	H int
}

// arg opt
var (
	delaytime        = flag.Int("delay", 10, "delay time: 100ths of a second")
	path      string = "./sample.jpg"
)

func decode(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ext := filepath.Ext(path)
	if ext == ".jpg" {
		return jpeg.Decode(file)
	} else if ext == ".png" {
		return png.Decode(file)
	}

	return nil, errors.New("unknown ext")
}

func createColor(decImg image.Image, bgcolor *image.Uniform) *color.Palette {
	qp := median.Quantizer(255).Quantize(make(color.Palette, 0, 255), decImg)
	mypalette := color.Palette{
		bgcolor,
	}
	for _, color := range qp {
		mypalette = append(mypalette, color)
	}
	return &mypalette
}

func newRGBA(wh WH, scale int, xy0 image.Point) *image.RGBA {
	f := func(p0, max int) int {
		offset := max * scale / 100
		if p0+offset > max {
			return max
		}
		return p0 + offset
	}
	return image.NewRGBA(image.Rect(xy0.X, xy0.Y, f(xy0.X, wh.W), f(xy0.Y, wh.H)))
}

func arrayImage(decImg image.Image, pos int, bgcolor *image.Uniform) *image.Paletted {
	rect := decImg.Bounds()
	startPos := []image.Point{{0, 0},
		{rect.Dx() / 2, 0},
		{0, rect.Dy() / 2},
		{rect.Dx() / 2, rect.Dy() / 2}}
	smallImg := newRGBA(WH{rect.Dx(), rect.Dy()}, 50, startPos[pos%len(startPos)])
	draw.CatmullRom.Scale(smallImg, smallImg.Bounds(), decImg, rect, draw.Over, nil)
	color := createColor(decImg, bgcolor)

	paletted := image.NewPaletted(rect, *color)
	smallrect := smallImg.Bounds()

	for x := smallrect.Min.X; x < smallrect.Max.X; x++ {
		for y := smallrect.Min.Y; y < smallrect.Max.Y; y++ {
			paletted.Set(x, y, smallImg.At(x, y))
		}
	}

	return paletted
}
func create(decImg image.Image) (*gif.GIF, error) {
	rect := decImg.Bounds()
	dst := gif.GIF{
		Image:     []*image.Paletted{},
		LoopCount: 0,
	}

	// mycolor := palette.WebSafe
	mycolor := createColor(decImg, image.Transparent)
	f := func(at func(x, y int) (int, int)) {
		paletted := image.NewPaletted(rect, *mycolor)
		for x := rect.Min.X; x < rect.Max.X; x++ {
			for y := rect.Min.Y; y < rect.Max.Y; y++ {
				paletted.Set(x, y, decImg.At(at(x, y)))
			}
		}
		dst.Image = append(dst.Image, paletted)
	}

	f(func(x, y int) (int, int) { return x, y }) // original
	f(func(x, y int) (int, int) { return -y + rect.Max.Y, x })
	f(func(x, y int) (int, int) { return -y + rect.Max.Y, -x + rect.Max.X })
	f(func(x, y int) (int, int) { return x, -y + rect.Max.X })

	dst.Image = append(dst.Image, arrayImage(decImg, 0, image.Black))
	dst.Image = append(dst.Image, arrayImage(decImg, 1, image.Black))
	dst.Image = append(dst.Image, arrayImage(decImg, 2, image.Black))
	dst.Image = append(dst.Image, arrayImage(decImg, 3, image.Black))
	dst.Image = append(dst.Image, arrayImage(decImg, 0, image.Black))
	dst.Image = append(dst.Image, arrayImage(decImg, 1, image.Transparent))
	dst.Image = append(dst.Image, arrayImage(decImg, 2, image.Transparent))
	dst.Image = append(dst.Image, arrayImage(decImg, 3, image.Transparent))

	dst.Delay = make([]int, len(dst.Image))
	for i := 0; i < len(dst.Delay); i++ {
		dst.Delay[i] = *delaytime
	}

	return &dst, nil
}

func save(gifImg *gif.GIF) error {
	name := path[0:len(path)-len(filepath.Ext(path))] + ".gif"
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return gif.EncodeAll(file, gifImg)
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) > 0 {
		path = flag.Args()[0]
	}

	decImg, err := decode(path)
	if err != nil {
		log.Fatal(err)
	}

	gifImg, err := create(decImg)
	if err != nil {
		log.Fatal(err)
	}

	err = save(gifImg)
	if err != nil {
		log.Fatal(err)
	}
}
