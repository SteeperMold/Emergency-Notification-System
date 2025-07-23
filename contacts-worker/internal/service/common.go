package service

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/phoneutils"
	"runtime"
	"sync"
	"sync/atomic"
)

type RowProvider func(jobsCh chan<- []string) error

func (cs *ContactsService) ingestAndSave(ctx context.Context, userID int, provider RowProvider) (int, error) {
	var total32 int32

	jobsCh := make(chan []string, runtime.NumCPU()*2)
	writeCh := make(chan []*models.Contact, runtime.NumCPU())

	var wgWriter sync.WaitGroup
	wgWriter.Add(1)
	go func() {
		defer wgWriter.Done()
		for batch := range writeCh {
			cs.saveBatch(ctx, &batch, &total32)
		}
	}()

	var wgWorkers sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wgWorkers.Add(1)
		go func() {
			defer wgWorkers.Done()
			batch := make([]*models.Contact, 0, cs.batchSize)
			for rec := range jobsCh {
				if len(rec) < 2 {
					continue
				}

				name, nameOk := cs.validateName(rec[0])
				phone, phoneOk := cs.validatePhone(rec[1])
				if !nameOk || !phoneOk {
					continue
				}

				batch = append(batch, &models.Contact{
					UserID: userID,
					Name:   name,
					Phone:  phone,
				})

				if len(batch) >= cs.batchSize {
					writeCh <- batch
					batch = make([]*models.Contact, 0, cs.batchSize)
				}
			}
			if len(batch) > 0 {
				writeCh <- batch
			}
		}()
	}

	err := provider(jobsCh)
	if err != nil {
		return 0, err
	}

	close(jobsCh)
	wgWorkers.Wait()

	close(writeCh)
	wgWriter.Wait()

	return int(atomic.LoadInt32(&total32)), nil
}

func (cs *ContactsService) saveBatch(ctx context.Context, batch *[]*models.Contact, total *int32) {
	err := cs.repository.SaveContacts(ctx, *batch)
	if err == nil { // if no error
		atomic.AddInt32(total, int32(len(*batch)))
	}
}

func (cs *ContactsService) validateName(name string) (string, bool) {
	return name, len(name) > 0 && len(name) <= 32
}

func (cs *ContactsService) validatePhone(phone string) (string, bool) {
	formattedNum, err := phoneutils.FormatToE164(phone, phoneutils.RegionRU)
	return formattedNum, err == nil
}
