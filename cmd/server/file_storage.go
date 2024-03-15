package main

import (
	"context"
	"encoding/json"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"io"
	"os"
	"sync"
	"time"
)

type FileStorage struct {
	file          *os.File
	encoder       *json.Encoder
	decoder       *json.Decoder
	needRestore   bool
	fileMutex     sync.Mutex
	storeInterval uint
}

func NewFileStorage(conf config) (*FileStorage, error) {
	file, err := os.OpenFile(conf.fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:          file,
		encoder:       json.NewEncoder(file),
		decoder:       json.NewDecoder(file),
		needRestore:   conf.restore,
		storeInterval: conf.storeInterval,
	}, nil
}

func (fs *FileStorage) Restore(ctx context.Context, st repository.Storage) error {
	fs.fileMutex.Lock()
	defer fs.fileMutex.Unlock()

	if !fs.needRestore {
		return nil
	}

	if _, err := fs.file.Seek(0, io.SeekStart); err != nil {
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

	if err := fs.file.Truncate(0); err != nil {
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

	if err := fs.file.Sync(); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) SyncByInterval(ctx context.Context, app *app, ch chan bool) error {
	if fs.storeInterval == 0 || app.memStorage == nil {
		close(ch)
		return nil
	}

	storeIntervalDuration := time.Duration(fs.storeInterval) * time.Second
	forever := make(chan bool)
	err := func() error {
		for {
			select {
			case <-forever:
				return nil
			default:
				<-time.After(storeIntervalDuration)
				if err := fs.Sync(ctx, app.memStorage); err != nil {
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
