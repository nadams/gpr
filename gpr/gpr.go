package gpr

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type GPR struct {
	Rows []Row
}

func (g *GPR) ByProtein() map[string][]Row {
	m := make(map[string][]Row)

	for _, row := range g.Rows {
		slc, ok := m[row.ID]
		if !ok {
			slc = make([]Row, 0, 2)
		}

		slc = append(slc, row)

		if len(slc) > 1 {
			sort.Slice(slc, func(i, j int) bool {
				return slc[i].X < slc[j].X && slc[i].Y < slc[j].Y
			})
		}

		m[row.ID] = slc
	}

	return m
}

func (g *GPR) SortByID() *GPR {
	sort.Slice(g.Rows, func(i, j int) bool {
		return g.Rows[i].ID < g.Rows[j].ID
	})

	return g
}

func (g *GPR) SortBy650Median() *GPR {
	sort.Slice(g.Rows, func(i, j int) bool {
		return g.Rows[i].F650MedianB650 > g.Rows[j].F650MedianB650
	})

	return g
}

func (g *GPR) SortBy550Median() *GPR {
	sort.Slice(g.Rows, func(i, j int) bool {
		return g.Rows[i].F550MedianB550 > g.Rows[j].F550MedianB550
	})

	return g
}

func (g *GPR) SortBy650Mean() *GPR {
	sort.Slice(g.Rows, func(i, j int) bool {
		return g.Rows[i].F650MeanB650 > g.Rows[j].F650MeanB650
	})

	return g
}

func (g *GPR) SortBy550Mean() *GPR {
	sort.Slice(g.Rows, func(i, j int) bool {
		return g.Rows[i].F550MeanB550 > g.Rows[j].F550MeanB550
	})

	return g
}

func (g *GPR) Averaged() *GPR {
	m := map[string][]Row{}

	for _, x := range g.Rows {
		m[x.ID] = append(m[x.ID], x)
	}

	n := make([]Row, 0, len(m))

	for _, v := range m {
		if len(v) != 2 {
			continue
		}

		n = append(n, Row{
			ID:             v[0].ID,
			F650MeanB650:   (v[0].F650MeanB650 + v[1].F650MeanB650) / 2,
			F550MeanB550:   (v[0].F550MeanB550 + v[1].F550MeanB550) / 2,
			F650MedianB650: (v[0].F650MedianB650 + v[1].F650MedianB650) / 2,
			F550MedianB550: (v[0].F550MedianB550 + v[1].F550MedianB550) / 2,
			SNR650:         (v[0].SNR650 + v[1].SNR650) / 2,
			SNR550:         (v[0].SNR550 + v[1].SNR550) / 2,
		})
	}

	return &GPR{Rows: n}
}

type Row struct {
	ID             string
	X              int
	Y              int
	Block          int
	Column         int
	Row            int
	Diameter       int
	F650Median     float64
	F550Median     float64
	F650MeanB650   float64
	F550MeanB550   float64
	F650MedianB650 float64
	F550MedianB550 float64
	SNR650         float64
	SNR550         float64
}

func Read(path string) (*GPR, error) {
	if !strings.HasSuffix(path, ".gpr") {
		return nil, errors.New("not a gpr file")
	}

	rows := make([]Row, 0, 10000)

	p, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer p.Close()

	r := csv.NewReader(p)
	r.Comma = '\t'
	r.FieldsPerRecord = -1

	var i int

	for {
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if i < 33 {
			i++
			continue
		}

		if len(line) != 56 {
			fmt.Println("invalid line length")
			continue
		}

		switch line[2] {
		case "17", "18":
			continue
		}

		switch strings.ToLower(line[4]) {
		case "empty", "blank":
			continue
		}

		blockStr := line[0]
		colStr := line[1]
		rowStr := line[2]
		protein := line[4]
		xStr, yStr, diaStr := line[5], line[6], line[7]
		f650MedianStr := line[8]
		f550MedianStr := line[20]
		f650MedianMinusStr := line[45]
		f550MedianMinusStr := line[46]
		f650MeanStr := line[47]
		f550MeanStr := line[48]
		snr650Str := line[51]
		snr550Str := line[52]

		block, err := strconv.Atoi(blockStr)
		if err != nil {
			return nil, err
		}

		column, err := strconv.Atoi(colStr)
		if err != nil {
			return nil, err
		}

		row, err := strconv.Atoi(rowStr)
		if err != nil {
			return nil, err
		}

		x, err := strconv.Atoi(xStr)
		if err != nil {
			return nil, err
		}

		y, err := strconv.Atoi(yStr)
		if err != nil {
			return nil, err
		}

		dia, err := strconv.Atoi(diaStr)
		if err != nil {
			return nil, err
		}

		f650Mean, err := strconv.ParseFloat(f650MeanStr, 64)
		if err != nil {
			return nil, err
		}

		f550Mean, err := strconv.ParseFloat(f550MeanStr, 64)
		if err != nil {
			return nil, err
		}

		f650Median, err := strconv.ParseFloat(f650MedianStr, 64)
		if err != nil {
			return nil, err
		}

		f550Median, err := strconv.ParseFloat(f550MedianStr, 64)
		if err != nil {
			return nil, err
		}

		f650MedianMinus, err := strconv.ParseFloat(f650MedianMinusStr, 64)
		if err != nil {
			return nil, err
		}

		f550MedianMinus, err := strconv.ParseFloat(f550MedianMinusStr, 64)
		if err != nil {
			return nil, err
		}

		snr650, err := strconv.ParseFloat(snr650Str, 64)
		if err != nil {
			return nil, err
		}

		snr550, err := strconv.ParseFloat(snr550Str, 64)
		if err != nil {
			return nil, err
		}

		if f650Mean < 0 {
			f650Mean = 1
		}

		if f550Mean < 0 {
			f550Mean = 1
		}

		if f650MedianMinus < 0 {
			f650MedianMinus = 1
		}

		if f550MedianMinus < 0 {
			f550MedianMinus = 1
		}

		if snr650 < 0 {
			snr650 = 0
		}

		if snr550 < 0 {
			snr550 = 0
		}

		rows = append(rows, Row{
			ID:             protein,
			Block:          block,
			Column:         column,
			Row:            row,
			X:              x,
			Y:              y,
			Diameter:       dia,
			F550Median:     f550Median,
			F650Median:     f650Median,
			F650MeanB650:   f650Mean,
			F550MeanB550:   f550Mean,
			F650MedianB650: f650MedianMinus,
			F550MedianB550: f550MedianMinus,
			SNR650:         snr650,
			SNR550:         snr550,
		})

		i++
	}

	return &GPR{
		Rows: rows,
	}, nil
}
