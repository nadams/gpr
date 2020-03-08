package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/andviro/go-libtiff/libtiff"
	"github.com/anthonynsimon/bild/adjust"
	"github.com/cheggaaa/pb"
	"github.com/fogleman/gg"
	"github.com/goki/freetype/truetype"
	"github.com/markbates/pkger"
	"golang.org/x/image/font"

	"gitlab.node-3.net/nadams/gpr/colr"
	"gitlab.node-3.net/nadams/gpr/gpr"
	"gitlab.node-3.net/nadams/gpr/gps"
)

const (
	paddingx    = 60
	paddingy    = 34
	rectpadding = 7
	textheight  = 16
)

var (
	noRegex = regexp.MustCompile(`No\.\s*\d+`)
)

type CLI struct {
	Dir      string   `arg:"" name:"dir" help:"Directory containing tiff and gpr files." type:"existingdir" default:"."`
	Proteins []string `name:"proteins" help:"List of proteins to get, get all if empty." optional:""`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	ctx.FatalIfErrorf(ctx.Validate())

	fis, err := ioutil.ReadDir(cli.Dir)
	ctx.FatalIfErrorf(err)

	ff, err := loadFont(16)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	defer ff.Close()

	var bar *pb.ProgressBar

	var newfis []os.FileInfo
	for _, fi := range fis {
		switch strings.ToLower(filepath.Ext(fi.Name())) {
		case ".tif", ".tiff":
			newfis = append(newfis, fi)
		}
	}

	selectedProteins := map[string]struct{}{}
	for _, p := range cli.Proteins {
		selectedProteins[p] = struct{}{}
	}

	for _, fi := range newfis {
		fi := fi
		if err := func() error {
			name := strings.TrimSuffix(filepath.Base(fi.Name()), filepath.Ext(fi.Name()))
			gprPath := filepath.Join(cli.Dir, name+".gpr")
			gpsPath := filepath.Join(cli.Dir, name+".gps")
			tiffPath := filepath.Join(cli.Dir, fi.Name())
			nmbr := noRegex.FindString(tiffPath)

			if _, err := os.Stat(gprPath); os.IsNotExist(err) {
				return fmt.Errorf("missing gpr file for '%s'", fi.Name())
			}

			if _, err := os.Stat(gpsPath); os.IsNotExist(err) {
				return fmt.Errorf("missing gps file for '%s'", fi.Name())
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

			brightness, contrast, err := gps.BrightnessContrast(gpsPath)
			if err != nil {
				return err
			}

			proteins := data.ByProtein()
			outdir := filepath.Join(cli.Dir, "results", name)

			if bar == nil {
				var c int64

				if len(selectedProteins) > 0 {
					c = int64(len(selectedProteins)) * int64(2) * int64(len(newfis))
				} else {
					c = int64(len(proteins)) * int64(2) * int64(len(newfis))
				}

				bar = pb.Start64(c)
			}

			tiff.Iter(func(n int) {
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

				_ = brightness
				_ = contrast

				s := newimg
				s = adjust.Brightness(s, 1+float64(brightness)*0.01)
				s = adjust.Contrast(s, float64(contrast)*0.0005)
				s = adjust.Gamma(s, 1.8)

				for protein, spots := range proteins {
					if len(selectedProteins) > 0 {
						if _, ok := selectedProteins[protein]; !ok {
							continue
						}
					}

					r1, r2 := spots[0].Diameter/2, spots[1].Diameter/2
					x1, y1 := (spots[0].X-r1)/10, (spots[0].Y-r1)/10
					x2, y2 := (spots[1].X+r2)/10, (spots[1].Y+r2)/10
					rect := image.Rect(x1-paddingx, y1-paddingy, x2+paddingx, y2+paddingy)

					x := copyImg(s.SubImage(rect).(*image.RGBA))
					ctx := gg.NewContextForRGBA(x)
					ctx.SetColor(color.White)
					ctx.SetFontFace(ff)
					ctx.DrawStringWrapped(strings.ReplaceAll(nmbr, " ", ""), 20, 6, 0, 0, float64(x.Bounds().Max.X), 1, gg.AlignCenter)

					if err := ctx.SavePNG(filepath.Join(dir, fmt.Sprintf("%s.png", protein))); err != nil {
						panic(err)
					}

					bar.Increment()
				}
			})

			return nil
		}(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if bar != nil {
		bar.Finish()
	}
}

func copyImg(src *image.RGBA) *image.RGBA {
	origRect := src.Bounds()
	newRect := image.Rect(0, 0, origRect.Dx(), origRect.Dy())
	dst := image.NewRGBA(newRect)

	var newx int

	for x := origRect.Min.X; x < origRect.Max.X; x++ {
		var newy int

		for y := origRect.Min.Y; y < origRect.Max.Y; y++ {
			dst.Set(newx, newy, src.At(x, y))

			newy++
		}

		newx++
	}

	return dst
}

func loadFont(points float64) (font.Face, error) {
	fontFile, err := pkger.Open("/fonts/Gotham-Book.ttf")
	if err != nil {
		return nil, err
	}

	defer fontFile.Close()

	b, err := ioutil.ReadAll(fontFile)
	if err != nil {
		return nil, err
	}

	f, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(f, &truetype.Options{
		Size:    points,
		Hinting: font.HintingFull,
	}), nil
}
