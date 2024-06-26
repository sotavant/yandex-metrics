// Package storage Пакет для работы с хранилищами
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
)

// FileStorage структура для работы с файловым хранилищем
type FileStorage struct {
	File          *os.File
	encoder       *json.Encoder
	decoder       *json.Decoder
	needRestore   bool
	fileMutex     sync.Mutex
	StoreInterval uint
}

// NewFileStorage инициализация файлового хранилища.
//
// Параметры:
//
//	fileStorage - путь к файлу-хранилищу
//	needRestore - нужно ли восстанавливать значения метрик из файла
//	storeInterval - интервал для сброса значений в файл
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

// Restore метод для восстановления значения из файла
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

// Sync сброс значения из хранилища в файл
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

		if err = fs.encoder.Encode(&m); err != nil {
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

// SyncByInterval сброс значения в файл с заданным интервалом. Если интервал не задан, то не синхронизируется.
func (fs *FileStorage) SyncByInterval(ctx context.Context, storage repository.Storage) error {
	storeIntervalDuration := time.Duration(fs.StoreInterval) * time.Second
	forever := make(chan bool)
	err := func() error {
		for {
			select {
			case <-forever:
				return nil
			default:
				<-time.After(storeIntervalDuration)
				fmt.Println("syncing metrics")
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
