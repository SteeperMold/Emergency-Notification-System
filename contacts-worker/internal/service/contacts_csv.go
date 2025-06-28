package service

import (
	"context"
	"encoding/csv"
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
)

func (cs *ContactsService) processCsv(ctx context.Context, userID int, file io.Reader) (int, error) {
	r := csv.NewReader(file)
	r.FieldsPerRecord = -1

	var total int32
	workers := runtime.NumCPU() * 2
	jobs := make(chan []string, workers*100)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			const batchSize = 100
			var batch []kafka.Message

			for rec := range jobs {
				if len(rec) < 2 {
					continue
				}
				formattedName, nameOk := cs.validateName(rec[0])
				formattedPhone, phoneOk := cs.validatePhone(rec[1])
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
					cs.flushBatch(ctx, &batch, &total)
				}
			}
			if len(batch) > 0 {
				cs.flushBatch(ctx, &batch, &total)
			}
		}()
	}

	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		jobs <- rec
	}
	close(jobs)
	wg.Wait()

	return int(atomic.LoadInt32(&total)), nil
}
