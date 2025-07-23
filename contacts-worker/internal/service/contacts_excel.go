package service

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
)

func (cs *ContactsService) createExcelRowProvider(file io.Reader) (RowProvider, error) {
	f, rowsIter, err := cs.getRowsFromExcel(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rowsIter.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	provider := func(jobs chan<- []string) error {
		for rowsIter.Next() {
			cols, err := rowsIter.Columns()
			if err != nil {
				continue
			}
			jobs <- cols
		}
		return rowsIter.Error()
	}

	return provider, err
}

func (cs *ContactsService) getRowsFromExcel(file io.Reader) (*excelize.File, *excelize.Rows, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, nil, err
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("no sheets")
	}

	rowsIter, err := f.Rows(sheets[0])
	if err != nil {
		return nil, nil, err
	}

	return f, rowsIter, nil
}
