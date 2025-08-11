package service

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/phoneutils"
)

// rowProvider emits CSV records (as string slices) onto the provided jobsCh.
type rowProvider func(jobsCh chan<- []string) error

// ingestAndSave reads rows via provider, validates & batches them, and writes to repository.
func (cs *ContactsService) ingestAndSave(ctx context.Context, userID int, provider rowProvider) (int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var total int32

	jobsCh := make(chan []string, runtime.NumCPU()*2)
	writeCh := make(chan []*models.Contact, runtime.NumCPU())
	errCh := make(chan error, 1)

	// start writer
	var wgWriter sync.WaitGroup
	wgWriter.Add(1)
	go func() {
		defer wgWriter.Done()
		cs.runWriter(ctx, writeCh, errCh, &total)
	}()

	// start workers
	var wgWorkers sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wgWorkers.Add(1)
		go func() {
			defer wgWorkers.Done()
			cs.runWorker(ctx, userID, jobsCh, writeCh)
		}()
	}

	// feed jobs
	if err := provider(jobsCh); err != nil {
		close(jobsCh)
		return 0, err
	}
	close(jobsCh)

	wgWorkers.Wait()
	close(writeCh)

	wgWriter.Wait()
	select {
	case saveErr := <-errCh:
		return int(atomic.LoadInt32(&total)), saveErr
	default:
		return int(atomic.LoadInt32(&total)), nil
	}
}

// runWriter consumes batches from writeCh and saves them via repository.
func (cs *ContactsService) runWriter(
	ctx context.Context,
	writeCh <-chan []*models.Contact,
	errCh chan<- error,
	total *int32,
) {
	for batch := range writeCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := cs.repository.SaveContacts(ctx, batch); err != nil {
			// publish first error only
			select {
			case errCh <- err:
			default:
			}
			return
		}

		atomic.AddInt32(total, int32(len(batch)))
	}
}

// runWorker reads raw records, validates them, and sends full batches to writeCh.
func (cs *ContactsService) runWorker(
	ctx context.Context,
	userID int,
	jobsCh <-chan []string,
	writeCh chan<- []*models.Contact,
) {
	batch := make([]*models.Contact, 0, cs.batchSize)
	flush := func() {
		if len(batch) > 0 {
			writeCh <- batch
			batch = make([]*models.Contact, 0, cs.batchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return

		case rec, ok := <-jobsCh:
			if !ok {
				flush()
				return
			}

			if c := cs.makeContact(userID, rec); c != nil {
				batch = append(batch, c)
				if len(batch) >= cs.batchSize {
					flush()
				}
			}
		}
	}
}

// makeContact applies validation to a raw record and returns a Contact or nil.
func (cs *ContactsService) makeContact(userID int, rec []string) *models.Contact {
	if len(rec) < 2 {
		return nil
	}
	name, okName := cs.validateName(rec[0])
	phone, okPhone := cs.validatePhone(rec[1])
	if !okName || !okPhone {
		return nil
	}
	return &models.Contact{
		UserID: userID,
		Name:   name,
		Phone:  phone,
	}
}

func (cs *ContactsService) validateName(name string) (string, bool) {
	return name, len(name) > 0 && len(name) <= 32
}

func (cs *ContactsService) validatePhone(phone string) (string, bool) {
	formattedNum, err := phoneutils.FormatToE164(phone, phoneutils.RegionRU)
	return formattedNum, err == nil
}
