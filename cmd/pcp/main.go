package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"gitlab.node-3.net/nadams/gpr/gpr"
)

type CLI struct {
	Dir         string `arg:"" name:"dir" help:"Directory containing gpr files." type:"existingdir" default:"."`
	UseSubtract bool   `arg:"" name:"use-subtract" help:"Use the subtraction fields." optional:""`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	ctx.FatalIfErrorf(ctx.Validate())

	fis, err := ioutil.ReadDir(cli.Dir)
	ctx.FatalIfErrorf(err)

	newFis := make([]os.FileInfo, 0, len(fis))
	for _, f := range fis {
		switch strings.ToLower(filepath.Ext(f.Name())) {
		case ".gpr":
			newFis = append(newFis, f)
		}
	}

	resultsDir := filepath.Join(cli.Dir, "pcp_results")

	for _, fis := range newFis {
		if err := func() error {
			gprPath := filepath.Join(cli.Dir, fis.Name())
			groupName := strings.TrimSuffix(fis.Name(), ".gpr")

			data, err := gpr.Read(gprPath)
			if err != nil {
				return fmt.Errorf("could not load gpr data: %w", err)
			}

			iggDir := filepath.Join(resultsDir, "IgG")
			igmDir := filepath.Join(resultsDir, "IgM")

			if err := os.MkdirAll(iggDir, 0755); err != nil {
				return fmt.Errorf("could not create IgG dir: %w", err)
			}

			if err := os.MkdirAll(igmDir, 0755); err != nil {
				return fmt.Errorf("could not create IgM dir: %w", err)
			}

			iggFile, err := os.Create(filepath.Join(iggDir, groupName+".csv"))
			if err != nil {
				return fmt.Errorf("could not create IgG file: %w", err)
			}

			defer iggFile.Close()

			igmFile, err := os.Create(filepath.Join(igmDir, groupName+".csv"))
			if err != nil {
				return fmt.Errorf("could not create IgM file: %w", err)
			}

			defer igmFile.Close()

			iggOut := csv.NewWriter(iggFile)
			igmOut := csv.NewWriter(igmFile)

			defer iggOut.Flush()
			defer igmOut.Flush()

			switch {
			case cli.UseSubtract:
				iggOut.Write([]string{"ID", "F550 Median - B550", "Block", "Column", "Row"})
				igmOut.Write([]string{"ID", "F650 Median - B650", "Block", "Column", "Row"})

				for _, row := range data.SortByID().Rows {
					iggOut.Write([]string{row.ID, fmt.Sprintf("%v", row.F550MedianB550), strconv.Itoa(row.Block), strconv.Itoa(row.Column), strconv.Itoa(row.Row)})
					igmOut.Write([]string{row.ID, fmt.Sprintf("%v", row.F650MedianB650), strconv.Itoa(row.Block), strconv.Itoa(row.Column), strconv.Itoa(row.Row)})
				}
			default:
				iggOut.Write([]string{"ID", "F550 Median", "Block", "Column", "Row"})
				igmOut.Write([]string{"ID", "F650 Median", "Block", "Column", "Row"})

				for _, row := range data.SortByID().Rows {
					iggOut.Write([]string{row.ID, fmt.Sprintf("%v", row.F550Median), strconv.Itoa(row.Block), strconv.Itoa(row.Column), strconv.Itoa(row.Row)})
					igmOut.Write([]string{row.ID, fmt.Sprintf("%v", row.F650Median), strconv.Itoa(row.Block), strconv.Itoa(row.Column), strconv.Itoa(row.Row)})
				}
			}

			return nil
		}(); err != nil {
			ctx.FatalIfErrorf(err)
		}
	}
}
