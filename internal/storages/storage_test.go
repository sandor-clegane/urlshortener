package storages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookUp(t *testing.T) {
	type want struct {
		value       string
		expectFound bool
	}

	tests := []struct {
		name    string
		storage map[string]string
		key     string
		want    want
	}{
		{
			storage: map[string]string{"id1": "http://ya.ru"},
			name:    "simple test 1",
			key:     "id1",
			want: want{
				value:       "http://ya.ru",
				expectFound: true,
			},
		},
		{
			storage: map[string]string{"id1": "http://ya.ru"},
			name:    "simple test 2",
			key:     "id2",
			want: want{
				value:       "",
				expectFound: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStorage()

			for k, v := range tt.storage {
				s.Insert(context.Background(), k, v, "some_user")
			}

			gotValue, ok := s.LookUp(context.Background(), tt.key)
			if tt.want.expectFound {
				assert.True(t, ok)
				assert.Equal(t, tt.want.value, gotValue)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	type want struct {
		values []string
	}
	tests := []struct {
		name    string
		storage map[string]string
		want    want
	}{
		{
			name:    "simple test 1",
			storage: map[string]string{"k1": "v1", "k2": "v2"},
			want: want{
				values: []string{"v1", "v2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStorage()

			i := 0
			for k, v := range tt.storage {
				s.Insert(context.Background(), k, v, "some_user")
				assert.Equal(t, tt.want.values[i], s.storage[k])
				i++
			}
		})
	}
}




