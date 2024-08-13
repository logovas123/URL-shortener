package storage

import "errors"

// интерфейсы будем реализовывать в месте использования
// здесь общий storage для разных хранилищ

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)
