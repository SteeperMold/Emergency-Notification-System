package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
	"github.com/xuri/excelize/v2"
)

func (cs *ContactsService) processExcel(ctx context.Context, userID int, file io.Reader) (total int, err error) {
	f, rowsIter, err := cs.getRowsFromExcel(file)
	defer func(f *excelize.File) {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}(f)
	defer func(rowsIter *excelize.Rows) {
		if cerr := rowsIter.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}(rowsIter)
	if err != nil {
		return 0, err
	}

	var count32 int32
	workers := runtime.NumCPU() * 2
	jobs := make(chan []string, workers*100)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			const batchSize = 100
			var batch []kafka.Message

			for cols := range jobs {
				if len(cols) < 2 {
					continue
				}
				formattedName, nameOk := cs.validateName(cols[0])
				formattedPhone, phoneOk := cs.validatePhone(cols[1])
				if !nameOk || !phoneOk {
					continue
				}
				value, err := cs.makeMessage(userID, formattedName, formattedPhone)
				if err != nil {
					continue
				}
				batch = append(batch, kafka.Message{
					Value: value,
				})
				if len(batch) >= batchSize {
					cs.flushBatch(ctx, &batch, &count32)
				}
			}
			if len(batch) > 0 {
				cs.flushBatch(ctx, &batch, &count32)
			}
		}()
	}

	for rowsIter.Next() {
		cols, err := rowsIter.Columns()
		if err != nil {
			continue
		}
		jobs <- cols
	}
	if err := rowsIter.Error(); err != nil {
		close(jobs)
		wg.Wait()
		return int(atomic.LoadInt32(&count32)), err
	}

	close(jobs)
	wg.Wait()
	return int(atomic.LoadInt32(&count32)), nil
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
