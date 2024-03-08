package main

import "github.com/go-chi/chi/v5"

type app struct {
	config     *config
	memStorage *MemStorage
	fs         *FileStorage
}

func initApp() (*app, error) {
	config := new(config)
	config.parseFlags()

	mem := NewMemStorage()
	fs, err := NewFileStorage(*config)

	if err != nil {
		panic(err)
	}

	if err = fs.Restore(mem); err != nil {
		panic(err)
	}

	return &app{
		config:     config,
		memStorage: mem,
		fs:         fs,
	}, nil
}

func (app *app) syncFs() {
	err := app.fs.Sync(app.memStorage)
	if err != nil {
		panic(err)
	}

	err = app.fs.file.Close()
	if err != nil {
		panic(err)
	}
}

func (app *app) initRouters() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(app))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(app))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(app))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(app))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(app))))

	return r
}
