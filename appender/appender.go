package appender

import (
	"fmt"
	"sync"
	"time"

	"github.com/tealeg/xlsx"
)

type Formula string

type ColAppender struct {
	m     sync.Mutex
	row   int
	col   int
	sheet *xlsx.Sheet
}

func NewColAppender(sheet *xlsx.Sheet) *ColAppender {
	return &ColAppender{
		sheet: sheet,
	}
}

func (r *ColAppender) NewCol() {
	r.m.Lock()
	defer r.m.Unlock()

	r.col++
	r.row = 0
}

func (r *ColAppender) Append(values ...interface{}) {
	r.m.Lock()
	defer r.m.Unlock()

	for _, value := range values {
		cell := r.sheet.Cell(r.row, r.col)

		switch x := value.(type) {
		case int:
			cell.SetInt(x)
		case int32:
			cell.SetInt(int(x))
		case int16:
			cell.SetInt(int(x))
		case int8:
			cell.SetInt(int(x))
		case int64:
			cell.SetInt64(x)
		case float32:
			cell.SetFloat(float64(x))
		case float64:
			cell.SetFloat(x)
		case bool:
			cell.SetBool(x)
		case string:
			cell.SetString(x)
		case time.Time:
			cell.SetDateTime(x)
		case Formula:
			cell.SetFormula(string(x))
		default:
			cell.SetString(fmt.Sprintf("%v", x))
		}

		r.row++
	}
}

type RowAppender struct {
	m     sync.Mutex
	row   int
	col   int
	sheet *xlsx.Sheet
}

func NewRowAppender(sheet *xlsx.Sheet) *RowAppender {
	return &RowAppender{
		sheet: sheet,
	}
}

func (r *RowAppender) NewRow() {
	r.m.Lock()
	defer r.m.Unlock()

	r.row++
	r.col = 0
}

func (r *RowAppender) Append(values ...interface{}) {
	r.m.Lock()
	defer r.m.Unlock()

	for _, value := range values {
		cell := r.sheet.Cell(r.row, r.col)

		switch x := value.(type) {
		case int:
			cell.SetInt(x)
		case int32:
			cell.SetInt(int(x))
		case int16:
			cell.SetInt(int(x))
		case int8:
			cell.SetInt(int(x))
		case int64:
			cell.SetInt64(x)
		case float32:
			cell.SetFloat(float64(x))
		case float64:
			cell.SetFloat(x)
		case bool:
			cell.SetBool(x)
		case string:
			cell.SetString(x)
		case time.Time:
			cell.SetDateTime(x)
		default:
			cell.SetString(fmt.Sprintf("%v", x))
		}

		r.col++
	}
}
