package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

type CLI struct {
	Files []string `arg:"" name:"files" help:"List of files to split." default:"*.png"`
	Out   string   `arg:"" name:"out" help:"Directory to write splitted images." default:"split_images"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	ctx.FatalIfErrorf(ctx.Validate())

	for _, filePath := range cli.Files {
		out, err := filepath.Glob(filePath)
		if err != nil {
		} else {
			for _, f := range out {
				if err := processImage(f, cli.Out); err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}
}

const (
	w = 150
	h = 72
)

func processImage(in, out string) error {
	protName := strings.TrimSuffix(filepath.Base(in), ".png")

	img, err := imgio.Open(in)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(out, protName), 0755); err != nil {
		return err
	}

	cropped := transform.Crop(img, image.Rect(15, 9, 15+w*5, 9+h*14))

	for i := 0; i < 5; i++ {
		for j := 0; j < 14; j++ {
			res := transform.Crop(cropped, image.Rect(i*w+15+1, j*h+9+1, i*w+w+15, j*h+h+9))
			if err := imgio.Save(filepath.Join(out, protName, fmt.Sprintf("%d_%d.png", i+1, j+1)), res, imgio.PNGEncoder()); err != nil {
				log.Println(err)
				continue
			}
		}
	}

	return nil
}
