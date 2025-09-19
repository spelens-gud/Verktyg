package istore

import "time"

type Store interface {
	Get(key string) (value string, err error)
	Set(key, value string) (err error)
	Add(key, value string) (err error)
	Delete(key string) (err error)
}

type FileStore interface {
	Delete(key string)

	Set(k string, x interface{}, d time.Duration)

	Get(k string) (interface{}, bool)

	LoadFile(fname string) error

	SaveFile(fname string) error
}
