package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
)

type FileStorage struct {
	File          *os.File
	encoder       *json.Encoder
	decoder       *json.Decoder
	needRestore   bool
	fileMutex     sync.Mutex
	StoreInterval uint
}

func NewFileStorage(fileStorage string, needRestore bool, storeInterval uint) (*FileStorage, error) {
	file, err := os.OpenFile(fileStorage, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		File:          file,
		encoder:       json.NewEncoder(file),
		decoder:       json.NewDecoder(file),
		needRestore:   needRestore,
		StoreInterval: storeInterval,
	}, nil
}

func (fs *FileStorage) Restore(ctx context.Context, st repository.Storage) error {
	fs.fileMutex.Lock()
	defer fs.fileMutex.Unlock()

	if !fs.needRestore {
		return nil
	}

	if _, err := fs.File.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for {
		var m internal.Metrics

		err := fs.decoder.Decode(&m)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if err = st.AddValue(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileStorage) Sync(ctx context.Context, st repository.Storage) error {
	fs.fileMutex.Lock()
	defer fs.fileMutex.Unlock()

	if err := fs.File.Truncate(0); err != nil {
		return err
	}

	gaugeValues, err := st.GetGauge(ctx)
	if err != nil {
		return err
	}

	for k, v := range gaugeValues {
		m := internal.Metrics{
			ID:    k,
			MType: internal.GaugeType,
			Delta: nil,
			Value: &v,
		}

		if err := fs.encoder.Encode(&m); err != nil {
			return err
		}
	}

	counters, err := st.GetCounters(ctx)
	if err != nil {
		return err
	}

	for k, v := range counters {
		m := internal.Metrics{
			ID:    k,
			MType: internal.CounterType,
			Delta: &v,
			Value: nil,
		}

		if err := fs.encoder.Encode(&m); err != nil {
			return err
		}
	}

	if err := fs.File.Sync(); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) SyncByInterval(ctx context.Context, storage repository.Storage, ch chan bool) error {
	if fs.StoreInterval == 0 {
		close(ch)
		return nil
	}

	storeIntervalDuration := time.Duration(fs.StoreInterval) * time.Second
	forever := make(chan bool)
	err := func() error {
		for {
			select {
			case <-forever:
				return nil
			default:
				<-time.After(storeIntervalDuration)
				if err := fs.Sync(ctx, storage); err != nil {
					return err
				}
			}
		}
	}()

	if err != nil {
		return err
	}

	<-forever

	return nil
}
