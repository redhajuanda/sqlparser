/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package protoutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/redhajuanda/sqlparser/dependencies/vt/proto/vttime"
)

func TestDurationFromProto(t *testing.T) {

	tests := []struct {
		name      string
		in        *vttime.Duration
		expected  time.Duration
		isOk      bool
		shouldErr bool
	}{
		{
			name:      "success",
			in:        &vttime.Duration{Seconds: 1000},
			expected:  time.Second * 1000,
			isOk:      true,
			shouldErr: false,
		},
		{
			name:      "nil value",
			in:        nil,
			expected:  0,
			isOk:      false,
			shouldErr: false,
		},
		{
			name: "error",
			in: &vttime.Duration{
				// This is the max allowed seconds for a durationpb, plus 1.
				Seconds: int64(10000*365.25*24*60*60) + 1,
			},
			expected:  0,
			isOk:      true,
			shouldErr: true,
		},
		{
			name: "nanoseconds",
			in: &vttime.Duration{
				Seconds: 1,
				Nanos:   500000000,
			},
			expected:  time.Second + 500*time.Millisecond,
			isOk:      true,
			shouldErr: false,
		},
		{
			name: "out of range nanoseconds",
			in: &vttime.Duration{
				Seconds: -1,
				Nanos:   500000000,
			},
			expected:  0,
			isOk:      true,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual, ok, err := DurationFromProto(tt.in)
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tt.isOk, ok, "expected (_, ok, _) = DurationFromProto; to be ok = %v", tt.isOk)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
			assert.Equal(t, tt.isOk, ok, "expected (_, ok, _) = DurationFromProto; to be ok = %v", tt.isOk)
		})
	}
}

func TestDurationToProto(t *testing.T) {

	tests := []struct {
		name     string
		in       time.Duration
		expected *vttime.Duration
	}{
		{
			name:     "success",
			in:       time.Second * 1000,
			expected: &vttime.Duration{Seconds: 1000},
		},
		{
			name:     "zero duration",
			in:       0,
			expected: &vttime.Duration{},
		},
		{
			name:     "nanoseconds",
			in:       time.Second + 500*time.Millisecond,
			expected: &vttime.Duration{Seconds: 1, Nanos: 500000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := DurationToProto(tt.in)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
