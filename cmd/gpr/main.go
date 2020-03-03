package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx"

	"gitlab.node-3.net/nadams/gpr/appender"
	"gitlab.node-3.net/nadams/gpr/gpr"
)

func main() {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fi, err := ioutil.ReadDir(d)
	if err != nil {
		panic(err)
	}

	for _, f := range fi {
		if !strings.HasSuffix(f.Name(), ".gpr") {
			continue
		}

		res, err := gpr.Read(filepath.Join(d, f.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		if err := outputDoc(f.Name()+".xlsx", res); err != nil {
			log.Println(err)
			continue
		}
	}
}

func outputDoc(path string, doc *gpr.GPR) error {
	sheets := []string{
		"Raw Data",
		"F650Mean-B650",
		"F650Medium-B650_SNR650",
		"F550Mean-B550",
		"F550Medium-B550_SNR550",
	}

	spreadsheet := xlsx.NewFile()
	avg := doc.Averaged()

	for i, name := range sheets {
		sheet, err := spreadsheet.AddSheet(name)
		if err != nil {
			return err
		}

		appender := appender.NewRowAppender(sheet)

		switch i {
		case 0:
			appender.Append("ID", "F650 Median - B650", "F550 Median - B550", "F650 Mean - B650", "F550 Mean - B550", "SNR 650", "SNR 550")
			appender.NewRow()

			for _, row := range doc.SortByID().Rows {
				appender.Append(row.ID, row.F650MedianB650, row.F550MedianB550, row.F650MeanB650, row.F550MeanB550, row.SNR650, row.SNR550)
				appender.NewRow()
			}
		case 1:
			appender.Append("ID", "F650 Mean - B650")
			appender.NewRow()

			for _, row := range avg.SortBy650Mean().Rows {
				appender.Append(row.ID, row.F650MeanB650)
				appender.NewRow()
			}
		case 2:
			appender.Append("ID", "F650 Medium - B650", "SNR 650")
			appender.NewRow()

			for _, row := range avg.SortBy650Median().Rows {
				appender.Append(row.ID, row.F650MedianB650, row.SNR650)
				appender.NewRow()
			}
		case 3:
			appender.Append("ID", "F550 Mean - B550")
			appender.NewRow()

			for _, row := range avg.SortBy550Mean().Rows {
				appender.Append(row.ID, row.F550MeanB550)
				appender.NewRow()
			}
		case 4:
			appender.Append("ID", "F550 Medium - B550", "SNR 550")
			appender.NewRow()

			for _, row := range avg.SortBy550Median().Rows {
				appender.Append(row.ID, row.F550MedianB550, row.SNR550)
				appender.NewRow()
			}
		}
	}

	return spreadsheet.Save(path)
}
