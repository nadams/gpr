package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/kong"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

type CLI struct {
	Files []string `arg:"" name:"files" help:"List of files to split." default:"*.html"`
	Out   string   `arg:"" name:"out" help:"Directory to write splitted images." default:"split_images"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	ctx.FatalIfErrorf(ctx.Validate())

	if err := os.MkdirAll(cli.Out, 0755); err != nil {
		log.Fatalln(err)
	}

	for _, filePath := range cli.Files {
		out, err := filepath.Glob(filePath)
		if err != nil {
			out = []string{filePath}
		}

		for _, f := range out {
			if err := processHTML(f, cli.Out); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func processHTML(in, out string) error {
	if !strings.HasSuffix(in, ".html") {
		return nil
	}

	f, err := os.Open(in)
	if err != nil {
		return err
	}

	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	var protein, group string
	doc.Find(".output_wrapper .output").First().Children().Each(func(i int, s *goquery.Selection) {
		if i%2 == 0 {
			for _, str := range strings.Split(s.Find("pre").Text(), "\n") {
				parts := strings.Split(str, ": ")
				switch parts[0] {
				case "protein":
					protein = parts[1]
				case "group":
					group = parts[1]
				}
			}
		} else {
			b64, ok := s.Find(".output_png img").Attr("src")
			if !ok {
				log.Println("could not get image data")
				return
			}

			i := strings.Index(b64, ",")
			dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64[i+1:]))

			if err := processImage(protein, group, filepath.Join(out, withoutExt(in)), dec); err != nil {
				log.Println(protein, group, err)
				return
			}
		}
	})

	return nil
}

const (
	tWidthLeft    = 15
	tHeightTop    = 9
	tWidthRight   = 8
	tHeightBottom = 14
	padding       = 1
	w             = 150
	h             = 72
)

func processImage(protein, group, out string, reader io.Reader) error {
	img, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(out, protein), 0755); err != nil {
		return err
	}

	totalWidth := ((img.Bounds().Dx() - tWidthLeft - tWidthRight) / (w + padding*2)) + 1
	totalHeight := ((img.Bounds().Dy() - tHeightTop - tHeightBottom) / (h + padding*2)) + 1

	for i := 0; i < totalWidth; i++ {
		for j := 0; j < totalHeight; j++ {
			x1 := i*w + tWidthLeft + padding
			y1 := j*h + tHeightTop + padding
			x2 := i*w + w + tWidthRight
			y2 := j*h + h + tHeightTop

			res := transform.Crop(img, image.Rect(x1, y1, x2, y2))
			if _, _, _, a := res.At(x1, y1).RGBA(); a >= 65535 {
				if err := imgio.Save(filepath.Join(out, protein, fmt.Sprintf("%s_%s_%d_%d.png", protein, group, i+1, j+1)), res, imgio.PNGEncoder()); err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}

	return nil
}

func withoutExt(in string) string {
	base := filepath.Base(in)
	idx := strings.LastIndex(base, filepath.Ext(base))
	if idx < 0 {
		return in
	}

	return in[:idx]
}
