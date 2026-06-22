package storage

import "testing"

func TestRepositoryInterface(t *testing.T) {
	var _ Repository = (*repository)(nil)
}
