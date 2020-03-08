package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	"github.com/tealeg/xlsx"

	"gitlab.node-3.net/nadams/gpr/appender"
	"gitlab.node-3.net/nadams/gpr/gpr"
)

type wvtype int

const (
	IgG wvtype = iota + 1
	IgM
)

func (w wvtype) String() string {
	switch w {
	case IgG:
		return "IgG"
	case IgM:
		return "IgM"
	default:
		return ""
	}
}

type samplegroup int

const (
	SHRA3 samplegroup = iota + 1
	SHRB2
	SHRA3CHR6B2
	IgHKO
	Normal
)

func (s samplegroup) String() string {
	switch s {
	case SHRA3:
		return "A340W-18W_Hits1"
	case SHRB2:
		return "B240W-18W_Hits2"
	case SHRA3CHR6B2:
		return "Chr6B240W-18W_Hits3"
	case IgHKO:
		return "IgHKO_Hits4"
	case Normal:
		return "Normal_Hits5"
	default:
		return ""
	}
}

type gprmap struct {
	left  string
	right string
	group samplegroup
}

var (
	files = []gprmap{
		{left: "1", right: "13", group: SHRA3},
		{left: "2", right: "14", group: SHRA3},
		{left: "3", right: "15", group: SHRA3},
		{left: "4", right: "16", group: SHRA3},
		{left: "5", right: "17", group: SHRA3},
		{left: "6", right: "18", group: SHRA3},
		{left: "7", right: "19", group: SHRB2},
		{left: "8", right: "20", group: SHRB2},
		{left: "9", right: "21", group: SHRB2},
		{left: "10", right: "22", group: SHRB2},
		{left: "11", right: "23", group: SHRB2},
		{left: "12", right: "24", group: SHRB2},
		{left: "25", right: "31", group: SHRA3CHR6B2},
		{left: "26", right: "32", group: SHRA3CHR6B2},
		{left: "27", right: "33", group: SHRA3CHR6B2},
		{left: "28", right: "34", group: SHRA3CHR6B2},
		{left: "29", right: "35", group: SHRA3CHR6B2},
		{left: "30", right: "36", group: SHRA3CHR6B2},
		{left: "55", group: IgHKO},
		{left: "56", group: Normal},
	}
)

type CLI struct {
	Dir string `arg:"" name:"dir" help:"Directory containing gpr files." type:"existingdir" optional:""`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	ctx.FatalIfErrorf(ctx.Validate())

	if cli.Dir == "" {
		cli.Dir = "."
	}

	ctx.FatalIfErrorf(work(&cli))
}

func work(cli *CLI) error {
	// IgG -> 550
	// IgM -> 650

	for _, wv := range []wvtype{IgG, IgM} {
		spreadsheet := xlsx.NewFile()
		sheets := map[samplegroup]*xlsx.Sheet{}
		appenders := map[samplegroup]*appender.ColAppender{}

		for _, m := range files {
			var newSheet bool

			sheet, ok := sheets[m.group]
			if !ok {
				s, err := spreadsheet.AddSheet(m.group.String())
				if err != nil {
					return fmt.Errorf("%s - %w", spew.Sdump(m), err)
				}

				sheet = s
				newSheet = true
				sheets[m.group] = sheet
			}

			apndr, ok := appenders[m.group]
			if !ok {
				apndr = appender.NewColAppender(sheet)
				appenders[m.group] = apndr
			}

			left, err := match(cli.Dir, m.left)
			if err != nil {
				return err
			}

			var leftGPR, rightGPR *gpr.GPR

			leftGPR, err = gpr.Read(left)
			if err != nil {
				return err
			}

			leftGPR = leftGPR.Averaged().SortByID()

			if m.right != "" {
				right, err := match(cli.Dir, m.right)
				if err != nil {
					return err
				}

				rightGPR, err = gpr.Read(right)
				if err != nil {
					return err
				}

				rightGPR = rightGPR.Averaged().SortByID()
			}

			if newSheet {
				apndr.Append("ID")

				for _, p := range leftGPR.Rows {
					apndr.Append(p.ID)
				}

				apndr.NewCol()
			}

			switch wv {
			case IgG:
				if rightGPR != nil {
					apndr.Append(fmt.Sprintf("F550 Medium - B550 (No.%s)", m.left))
					for _, v := range leftGPR.Rows {
						apndr.Append(v.F550MedianB550)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("F550 Medium - B550 (No.%s)", m.right))
					for _, v := range rightGPR.Rows {
						apndr.Append(v.F550MedianB550)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("No.%s-No.%s", m.left, m.right))
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F550MedianB550 - rightGPR.Rows[i].F550MedianB550)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("No.%s/No.%s", m.left, m.right))
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F550MedianB550 / rightGPR.Rows[i].F550MedianB550)
					}

					apndr.NewCol()
				} else {
					apndr.Append("F550 Medium - B550")
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F550MedianB550)
					}

					apndr.NewCol()
				}
			case IgM:
				if rightGPR != nil {
					apndr.Append(fmt.Sprintf("F650 Medium - B650 (No.%s)", m.left))
					for _, v := range leftGPR.Rows {
						apndr.Append(v.F650MedianB650)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("F650 Medium - B650 (No.%s)", m.right))
					for _, v := range rightGPR.Rows {
						apndr.Append(v.F650MedianB650)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("No.%s-No.%s", m.left, m.right))
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F650MedianB650 - rightGPR.Rows[i].F650MedianB650)
					}

					apndr.NewCol()

					apndr.Append(fmt.Sprintf("No.%s/No.%s", m.left, m.right))
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F650MedianB650 / rightGPR.Rows[i].F650MedianB650)
					}

					apndr.NewCol()
				} else {
					apndr.Append("F650 Medium - B650")
					for i := range leftGPR.Rows {
						apndr.Append(leftGPR.Rows[i].F650MedianB650)
					}

					apndr.NewCol()
				}
			}
		}

		if err := func() error {
			f, err := os.Create(filepath.Join(cli.Dir, fmt.Sprintf("%s Results.xlsx", wv)))
			if err != nil {
				return err
			}

			defer f.Close()

			return spreadsheet.Write(f)
		}(); err != nil {
			return err
		}
	}

	return nil
}

func match(dir, part string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, fmt.Sprintf("*No.%s.gpr", part)))
	if err != nil {
		return "", err
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("glob did not match exactly one file: %s", spew.Sdump(matches))
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	matches, err = filepath.Glob(filepath.Join(dir, fmt.Sprintf("*No. %s.gpr", part)))
	if err != nil {
		return "", err
	}

	if len(matches) != 1 {
		return "", fmt.Errorf("glob did not match exactly one file: %s", spew.Sdump(matches))
	}

	return matches[0], nil
}
