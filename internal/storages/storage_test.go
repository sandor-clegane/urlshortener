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
			s, _ := NewInMemoryStorage()

			for k, v := range tt.storage {
				s.Insert(context.Background(), k, v, "some_user")
			}

			gotValue, err := s.LookUp(context.Background(), tt.key)
			if tt.want.expectFound {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.value, gotValue)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	type pair struct {
		first  string
		second string
	}
	type want struct {
		values []string
	}
	tests := []struct {
		name    string
		storage []pair
		want    want
	}{
		{
			name:    "simple test 1",
			storage: []pair{{"k1", "v1"}, {"k2", "v2"}},
			want: want{
				values: []string{"v1", "v2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := NewInMemoryStorage()

			for i, p := range tt.storage {
				s.Insert(context.Background(), p.first, p.second, "some_user")
				assert.Equal(t, tt.want.values[i], s.storage[p.first])
			}
		})
	}
}
