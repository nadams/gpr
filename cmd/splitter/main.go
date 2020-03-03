package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/andviro/go-libtiff/libtiff"
	"github.com/anthonynsimon/bild/adjust"

	"gitlab.node-3.net/nadams/gpr/colr"
	"gitlab.node-3.net/nadams/gpr/gpr"
)

const (
	padding = 50
)

type CLI struct {
	Dir string `arg:"" name:"dir" help:"Directory containing tiff and gpr files." type:"existingdir" optional:""`
	//Proteins []string `name:"proteins" help:"List of proteins to get, get all if empty."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	ctx.FatalIfErrorf(ctx.Validate())

	if cli.Dir == "" {
		cli.Dir = "."
	}

	fis, err := ioutil.ReadDir(cli.Dir)
	ctx.FatalIfErrorf(err)

	for _, fi := range fis {
		fi := fi
		if err := func() error {
			switch strings.ToLower(filepath.Ext(fi.Name())) {
			case ".tif", ".tiff":
			default:
				return nil
			}

			name := strings.TrimSuffix(filepath.Base(fi.Name()), filepath.Ext(fi.Name()))
			gprPath := filepath.Join(cli.Dir, name+".gpr")
			tiffPath := filepath.Join(cli.Dir, fi.Name())

			if _, err := os.Stat(gprPath); os.IsNotExist(err) {
				return fmt.Errorf("missing gpr file for '%s'", fi.Name())
			}

			tiff, err := libtiff.Open(tiffPath)
			if err != nil {
				return err
			}

			defer tiff.Close()

			data, err := gpr.Read(gprPath)
			if err != nil {
				return err
			}

			proteins := data.ByProtein()
			outdir := filepath.Join(cli.Dir, "results", name)

			n := tiff.Iter(func(n int) {
				img, err := tiff.GetRGBA()
				if err != nil {
					panic(err)
				}

				newimg := image.NewRGBA(img.Bounds())
				w, h := newimg.Bounds().Max.X, newimg.Bounds().Max.Y

				const M = 1<<16 - 1
				var m color.Model
				var t string

				switch n {
				case 0:
					m = colr.MonoRed64Model
					t = "IgM"
				case 1:
					m = colr.MonoGreen64Model
					t = "IgG"
				default:
					return
				}

				dir := filepath.Join(outdir, t)
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Println(err)
					return
				}

				for x := 0; x < w; x++ {
					for y := 0; y < h; y++ {
						mr64 := m.Convert(img.At(x, y))
						newimg.SetRGBA(x, y, color.RGBAModel.Convert(mr64).(color.RGBA))
					}
				}

				for protein, spots := range proteins {
					r1, r2 := spots[0].Diameter/2, spots[1].Diameter/2
					x1, y1 := (spots[0].X-r1)/10, (spots[0].Y-r1)/10
					x2, y2 := (spots[1].X+r2)/10, (spots[1].Y+r2)/10
					subimg := adjust.Gamma(newimg.SubImage(image.Rect(x1-padding, y1-padding, x2+padding, y2+padding)).(*image.RGBA), 1.5)

					if err := func() error {
						f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s.png", protein)))
						if err != nil {
							return err
						}

						defer f.Close()

						return png.Encode(f, subimg)
					}(); err != nil {
						panic(err)
					}
				}
			})

			fmt.Printf("Total pages: %d\n", n)

			return nil
		}(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
