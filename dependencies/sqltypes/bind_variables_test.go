/*
Copyright 2019 The Vitess Authors.

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

package sqltypes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	querypb "github.com/redhajuanda/sqlparser/dependencies/vt/proto/query"
)

// TestProtoConversions checks coverting to and fro between querypb.Value and sqltypes.Value.
func TestProtoConversions(t *testing.T) {
	tcases := []struct {
		name     string
		val      Value
		protoVal *querypb.Value
	}{
		{
			name:     "integer value",
			val:      TestValue(Int64, "1"),
			protoVal: &querypb.Value{Type: Int64, Value: []byte("1")},
		}, {
			name: "tuple value",
			val:  TestTuple(TestValue(VarChar, "1"), TestValue(Int64, "3")),
		}, {
			name: "tuple of tuple as a value",
			val: TestTuple(
				TestTuple(
					TestValue(VarChar, "1"),
					TestValue(Int64, "3"),
				),
				TestValue(Int64, "5"),
			),
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			got := ValueToProto(tcase.val)
			// If we have an expected protoVal, check that serialization matches.
			// For nested tuples, we do not attempt to generate a protoVal, as it is binary data.
			// We simply check that the roundtrip is correct.
			if tcase.protoVal != nil {
				require.True(t, proto.Equal(got, tcase.protoVal), "ValueToProto: %v, want %v", got, tcase.protoVal)
			}
			gotback := ProtoToValue(got)
			require.EqualValues(t, tcase.val, gotback)
		})
	}
}

func TestBuildBindVariables(t *testing.T) {
	tcases := []struct {
		in  map[string]any
		out map[string]*querypb.BindVariable
		err string
	}{{
		in:  nil,
		out: nil,
	}, {
		in: map[string]any{
			"k": int64(1),
		},
		out: map[string]*querypb.BindVariable{
			"k": Int64BindVariable(1),
		},
	}, {
		in: map[string]any{
			"k": byte(1),
		},
		err: "k: type uint8 not supported as bind var: 1",
	}}
	for _, tcase := range tcases {
		bindVars, err := BuildBindVariables(tcase.in)
		if tcase.err == "" {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, tcase.err)
		}
		if !BindVariablesEqual(bindVars, tcase.out) {
			t.Errorf("MapToBindVars(%v): %v, want %s", tcase.in, bindVars, tcase.out)
		}
	}
}

func TestBuildBindVariable(t *testing.T) {
	tcases := []struct {
		in  any
		out *querypb.BindVariable
		err string
	}{{
		in: "aa",
		out: &querypb.BindVariable{
			Type:  querypb.Type_VARCHAR,
			Value: []byte("aa"),
		},
	}, {
		in: []byte("aa"),
		out: &querypb.BindVariable{
			Type:  querypb.Type_VARBINARY,
			Value: []byte("aa"),
		},
	}, {
		in: true,
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT8,
			Value: []byte("1"),
		},
	}, {
		in: false,
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT8,
			Value: []byte("0"),
		},
	}, {
		in: int(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}, {
		in: uint(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_UINT64,
			Value: []byte("1"),
		},
	}, {
		in: int32(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT32,
			Value: []byte("1"),
		},
	}, {
		in: uint32(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_UINT32,
			Value: []byte("1"),
		},
	}, {
		in: int64(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}, {
		in: uint64(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_UINT64,
			Value: []byte("1"),
		},
	}, {
		in: float64(1),
		out: &querypb.BindVariable{
			Type:  querypb.Type_FLOAT64,
			Value: []byte("1"),
		},
	}, {
		in:  nil,
		out: NullBindVariable,
	}, {
		in: MakeTrusted(Int64, []byte("1")),
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
		out: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}, {
		in: []any{"aa", int64(1)},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_VARCHAR,
				Value: []byte("aa"),
			}, {
				Type:  querypb.Type_INT64,
				Value: []byte("1"),
			}},
		},
	}, {
		in: []string{"aa", "bb"},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_VARCHAR,
				Value: []byte("aa"),
			}, {
				Type:  querypb.Type_VARCHAR,
				Value: []byte("bb"),
			}},
		},
	}, {
		in: [][]byte{[]byte("aa"), []byte("bb")},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_VARBINARY,
				Value: []byte("aa"),
			}, {
				Type:  querypb.Type_VARBINARY,
				Value: []byte("bb"),
			}},
		},
	}, {
		in: []int{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_INT64,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_INT64,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []uint{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_UINT64,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_UINT64,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []int32{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_INT32,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_INT32,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []uint32{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_UINT32,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_UINT32,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []int64{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_INT64,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_INT64,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []uint64{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_UINT64,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_UINT64,
				Value: []byte("2"),
			}},
		},
	}, {
		in: []float64{1, 2},
		out: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_FLOAT64,
				Value: []byte("1"),
			}, {
				Type:  querypb.Type_FLOAT64,
				Value: []byte("2"),
			}},
		},
	}, {
		in:  byte(1),
		err: "type uint8 not supported as bind var: 1",
	}, {
		in:  []any{1, byte(1)},
		err: "type uint8 not supported as bind var: 1",
	}}
	for _, tcase := range tcases {
		t.Run(fmt.Sprintf("%v", tcase.in), func(t *testing.T) {
			bv, err := BuildBindVariable(tcase.in)
			if tcase.err != "" {
				require.EqualError(t, err, tcase.err)
			} else {
				require.Truef(t, proto.Equal(tcase.out, bv), "binvar output did not match")
			}
		})
	}
}

func TestValidateBindVarables(t *testing.T) {
	tcases := []struct {
		in  map[string]*querypb.BindVariable
		err string
	}{{
		in: map[string]*querypb.BindVariable{
			"v": {
				Type:  querypb.Type_INT64,
				Value: []byte("1"),
			},
		},
		err: "",
	}, {
		in: map[string]*querypb.BindVariable{
			"v": {
				Type:  querypb.Type_INT64,
				Value: []byte("a"),
			},
		},
		err: `v: cannot parse int64 from "a"`,
	}, {
		in: map[string]*querypb.BindVariable{
			"v": {
				Type: querypb.Type_TUPLE,
				Values: []*querypb.Value{{
					Type:  Int64,
					Value: []byte("a"),
				}},
			},
		},
		err: `v: cannot parse int64 from "a"`,
	}}
	for _, tcase := range tcases {
		err := ValidateBindVariables(tcase.in)
		if tcase.err != "" {
			assert.ErrorContains(t, err, tcase.err)
			continue
		}
		assert.NoError(t, err)
	}
}

func TestValidateBindVariable(t *testing.T) {
	testcases := []struct {
		in  *querypb.BindVariable
		err string
	}{{
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT8,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT16,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT24,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT32,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT8,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT16,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT24,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT32,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT64,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_FLOAT32,
			Value: []byte("1.00"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_FLOAT64,
			Value: []byte("1.00"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_DECIMAL,
			Value: []byte("1.00"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_TIMESTAMP,
			Value: []byte("2012-02-24 23:19:43"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_DATE,
			Value: []byte("2012-02-24"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_TIME,
			Value: []byte("23:19:43"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_DATETIME,
			Value: []byte("2012-02-24 23:19:43"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_YEAR,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_TEXT,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_BLOB,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_VARCHAR,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_BINARY,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_CHAR,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_BIT,
			Value: []byte("1"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_ENUM,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_SET,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_VARBINARY,
			Value: []byte("a"),
		},
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte(InvalidNeg),
		},
		err: `cannot parse int64 from "-9223372036854775809": overflow`,
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_INT64,
			Value: []byte(InvalidPos),
		},
		err: `cannot parse int64 from "18446744073709551616": overflow`,
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT64,
			Value: []byte("-1"),
		},
		err: `cannot parse uint64 from "-1"`,
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_UINT64,
			Value: []byte(InvalidPos),
		},
		err: `cannot parse uint64 from "18446744073709551616": overflow`,
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_FLOAT64,
			Value: []byte("a"),
		},
		err: `unparsed tail left after parsing float64 from "a"`,
	}, {
		in: &querypb.BindVariable{
			Type:  querypb.Type_EXPRESSION,
			Value: []byte("a"),
		},
		err: "invalid type specified for MakeValue: EXPRESSION",
	}, {
		in: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type:  querypb.Type_INT64,
				Value: []byte("1"),
			}},
		},
	}, {
		in: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
		},
		err: "empty tuple is not allowed",
	}, {
		in: &querypb.BindVariable{
			Type: querypb.Type_TUPLE,
			Values: []*querypb.Value{{
				Type: querypb.Type_TUPLE,
			}},
		},
		err: "tuple not allowed inside another tuple",
	}}
	for _, tcase := range testcases {
		err := ValidateBindVariable(tcase.in)
		if tcase.err != "" {
			assert.ErrorContains(t, err, tcase.err)
			continue
		}
		assert.NoError(t, err)
	}

	// Special case: nil bind var.
	err := ValidateBindVariable(nil)
	want := "bind variable is nil"
	assert.ErrorContains(t, err, want)
}

func TestBindVariableToValue(t *testing.T) {
	v, err := BindVariableToValue(Int64BindVariable(1))
	require.NoError(t, err)
	assert.Equal(t, MakeTrusted(querypb.Type_INT64, []byte("1")), v)

	_, err = BindVariableToValue(&querypb.BindVariable{Type: querypb.Type_TUPLE})
	require.EqualError(t, err, "cannot convert a TUPLE bind var into a value")

	v, err = BindVariableToValue(BitNumBindVariable([]byte("0b101")))
	require.NoError(t, err)
	assert.Equal(t, MakeTrusted(querypb.Type_BITNUM, []byte("0b101")), v)

}

func TestBindVariablesEqual(t *testing.T) {
	bv1 := map[string]*querypb.BindVariable{
		"k": {
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}
	bv2 := map[string]*querypb.BindVariable{
		"k": {
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}
	bv3 := map[string]*querypb.BindVariable{
		"k": {
			Type:  querypb.Type_INT64,
			Value: []byte("1"),
		},
	}
	assert.True(t, BindVariablesEqual(bv1, bv2))
	assert.True(t, BindVariablesEqual(bv1, bv3))
}

func TestBindVariablesFormat(t *testing.T) {
	tupleBindVar, err := BuildBindVariable([]int64{1, 2})
	require.NoError(t, err, "failed to create a tuple bind var")

	bindVariables := map[string]*querypb.BindVariable{
		"key_1": StringBindVariable("val_1"),
		"key_2": Int64BindVariable(789),
		"key_3": BytesBindVariable([]byte("val_3")),
		"key_4": tupleBindVar,
	}

	formattedStr := FormatBindVariables(bindVariables, true /* full */, false /* asJSON */)
	assert.Contains(t, formattedStr, "key_1")
	assert.Contains(t, formattedStr, "val_1")

	assert.Contains(t, formattedStr, "key_2")
	assert.Contains(t, formattedStr, "789")

	assert.Contains(t, formattedStr, "key_3")
	assert.Contains(t, formattedStr, "val_3")

	assert.Contains(t, formattedStr, "key_4:type:TUPLE")

	formattedStr = FormatBindVariables(bindVariables, false /* full */, false /* asJSON */)
	assert.Contains(t, formattedStr, "key_1")

	assert.Contains(t, formattedStr, "key_2")
	assert.Contains(t, formattedStr, "789")

	assert.Contains(t, formattedStr, "key_3")
	assert.Contains(t, formattedStr, "5 bytes")

	assert.Contains(t, formattedStr, "key_4")
	assert.Contains(t, formattedStr, "2 items")

	formattedStr = FormatBindVariables(bindVariables, true /* full */, true /* asJSON */)
	assert.Contains(t, formattedStr, "\"key_1\": {\"type\": \"VARCHAR\", \"value\": \"val_1\"}")
	assert.Contains(t, formattedStr, "\"key_2\": {\"type\": \"INT64\", \"value\": 789}")
	assert.Contains(t, formattedStr, "\"key_3\": {\"type\": \"VARBINARY\", \"value\": \"val_3\"}")
	assert.Contains(t, formattedStr, "\"key_4\": {\"type\": \"TUPLE\", \"value\": \"\"}")

	formattedStr = FormatBindVariables(bindVariables, false /* full */, true /* asJSON */)
	assert.Contains(t, formattedStr, "\"key_1\": {\"type\": \"VARCHAR\", \"value\": \"5 bytes\"}")
	assert.Contains(t, formattedStr, "\"key_2\": {\"type\": \"INT64\", \"value\": 789}")
	assert.Contains(t, formattedStr, "\"key_3\": {\"type\": \"VARCHAR\", \"value\": \"5 bytes\"}")
	assert.Contains(t, formattedStr, "\"key_4\": {\"type\": \"VARCHAR\", \"value\": \"2 items\"}")
}
