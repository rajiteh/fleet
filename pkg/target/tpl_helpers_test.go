package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenParsing(t *testing.T) {

	cc := NewTplConversionCtx()
	cc.nonce = "123"

	tests := []struct {
		name         string
		inputVal     string
		unwrappedVal any
		isWrapped    bool
		panics       bool
	}{
		// sanity
		{
			name:     "wrappingIncorrectNonce",
			inputVal: "fleetYamlTplTypeConv:999:bool:true",
			panics:   true,
		},
		{
			name:         "wrappingIncorrectPrefix",
			inputVal:     "otherPrefix:123:bool:true",
			unwrappedVal: "otherPrefix:123:bool:true",
			isWrapped:    false,
		},
		{
			name:         "wrappedTokenBool",
			inputVal:     "fleetYamlTplTypeConv:123:bool:true",
			unwrappedVal: true,
			isWrapped:    true,
		},
		{
			name:         "wrappedTokenEmpty",
			inputVal:     "fleetYamlTplTypeConv:123:bool:",
			unwrappedVal: false,
			isWrapped:    true,
		},
		{
			name:         "incorrectWrapMissingSep",
			inputVal:     "fleetYamlTplTypeConv:123:bool",
			unwrappedVal: "fleetYamlTplTypeConv:123:bool",
			isWrapped:    false,
		},
		{
			name:         "icorrectWrapMissingKind",
			inputVal:     "fleetYamlTplTypeConv:123:",
			unwrappedVal: "fleetYamlTplTypeConv:123:",
			isWrapped:    false,
		},
		{
			name:         "regular string",
			inputVal:     "abcdef",
			unwrappedVal: "abcdef",
			isWrapped:    false,
		},
		{
			name:         "empty string",
			inputVal:     "",
			unwrappedVal: "",
			isWrapped:    false,
		},
		{
			name:         "special characters",
			inputVal:     "::??//\\;'[]!@#$%^&*(",
			unwrappedVal: "::??//\\;'[]!@#$%^&*(",
			isWrapped:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tplTypedToken(tt.inputVal)
			if tt.panics {
				assert.Panics(t, func() { token.IsWrapped(&cc) })
				return
			}
			assert.Equal(t, tt.isWrapped, token.IsWrapped(&cc), "unexpected wrapped status")
			assert.Equal(t, tt.inputVal, token.String(), "unexpected String() value")
			assert.Equal(t, tt.unwrappedVal, token.Unwrap(&cc), "unexpected Unwrap() value")
		})
	}

}

func TestTokenWrapping(t *testing.T) {

	cc := NewTplConversionCtx()
	cc.nonce = "123"

	tests := []struct {
		name         string
		valueType    tplValueType
		inputVal     string
		wrappedVal   string
		unwrappedVal any
		unwrapPanics bool
	}{

		// asInt
		{
			name:         "asInt",
			valueType:    tplValueTypeInt,
			inputVal:     "91919",
			wrappedVal:   "fleetYamlTplTypeConv:123:int:91919",
			unwrappedVal: int64(91919),
		},
		{
			name:         "asIntInvalid",
			valueType:    tplValueTypeInt,
			inputVal:     "91.919",
			wrappedVal:   "fleetYamlTplTypeConv:123:int:91.919",
			unwrapPanics: true,
		},
		{
			name:         "asIntInvalidNotANumber",
			valueType:    tplValueTypeInt,
			inputVal:     "91abcd",
			wrappedVal:   "fleetYamlTplTypeConv:123:int:91abcd",
			unwrapPanics: true,
		},
		{
			name:         "asIntInvalidEmpty",
			valueType:    tplValueTypeInt,
			inputVal:     "",
			wrappedVal:   "fleetYamlTplTypeConv:123:int:",
			unwrapPanics: true,
		},
		// asFloat
		{
			name:         "asFloat",
			valueType:    tplValueTypeFloat,
			inputVal:     "919.19",
			wrappedVal:   "fleetYamlTplTypeConv:123:float:919.19",
			unwrappedVal: float64(919.19),
		},
		{
			name:         "asFloatFromInt",
			valueType:    tplValueTypeFloat,
			inputVal:     "919",
			wrappedVal:   "fleetYamlTplTypeConv:123:float:919",
			unwrappedVal: float64(919.0),
		},
		{
			name:         "asFloatInvalidFloat",
			valueType:    tplValueTypeFloat,
			inputVal:     "9.1.9",
			wrappedVal:   "fleetYamlTplTypeConv:123:float:9.1.9",
			unwrapPanics: true,
		},
		{
			name:         "asFloatInvalidEmpty",
			valueType:    tplValueTypeFloat,
			inputVal:     "",
			wrappedVal:   "fleetYamlTplTypeConv:123:float:",
			unwrapPanics: true,
		},
		// asBool
		{
			name:         "asBoolTrueString",
			valueType:    tplValueTypeBool,
			inputVal:     "true",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:true",
			unwrappedVal: true,
		},
		{
			name:         "asBoolTrueStringVal",
			valueType:    tplValueTypeBool,
			inputVal:     "someValue",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:someValue",
			unwrappedVal: true,
		},
		{
			name:         "asBoolFalseString",
			valueType:    tplValueTypeBool,
			inputVal:     "FALSe",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:FALSe",
			unwrappedVal: false,
		},
		{
			name:         "asBoolZeroString",
			valueType:    tplValueTypeBool,
			inputVal:     "0",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:0",
			unwrappedVal: false,
		},
		{
			name:         "asBoolEmptyString",
			valueType:    tplValueTypeBool,
			inputVal:     "",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:",
			unwrappedVal: false,
		},
		{
			name:         "asBoolWhitespaceString",
			valueType:    tplValueTypeBool,
			inputVal:     "       ",
			wrappedVal:   "fleetYamlTplTypeConv:123:bool:       ",
			unwrappedVal: false,
		},
		// asNullable
		{
			name:         "asNullableEmptyString",
			valueType:    tplValueTypeNullable,
			inputVal:     "",
			wrappedVal:   "fleetYamlTplTypeConv:123:nullable:",
			unwrappedVal: nil,
		},
		{
			name:         "asNullableNilString",
			valueType:    tplValueTypeNullable,
			inputVal:     "nil",
			wrappedVal:   "fleetYamlTplTypeConv:123:nullable:nil",
			unwrappedVal: nil,
		},
		{
			name:         "asNullableNullString",
			valueType:    tplValueTypeNullable,
			inputVal:     "null",
			wrappedVal:   "fleetYamlTplTypeConv:123:nullable:null",
			unwrappedVal: nil,
		},
		{
			name:         "asNullableHasValue",
			valueType:    tplValueTypeNullable,
			inputVal:     "abc",
			wrappedVal:   "fleetYamlTplTypeConv:123:nullable:abc",
			unwrappedVal: "abc",
		},
		{
			name:         "asNullableHasWhitespace",
			valueType:    tplValueTypeNullable,
			inputVal:     "  ",
			wrappedVal:   "fleetYamlTplTypeConv:123:nullable:  ",
			unwrappedVal: "  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unwrapped := tplTypedToken(tt.inputVal)
			assert.Equal(t, tt.inputVal, unwrapped.String(), "unexpected unwrapped String() value")

			wrapped := unwrapped.Wrap(tt.valueType, cc)
			assert.Equal(t, tt.wrappedVal, wrapped.String(), "unexpected wrapped String() value")

			if tt.unwrapPanics {
				assert.Panics(t, func() { wrapped.Unwrap(&cc) })
				return
			}

			assert.Equal(t, tt.unwrappedVal, wrapped.Unwrap(&cc), "unexpected Unwrap() value")

		})
	}

}

func TestConversionFunctions(t *testing.T) {
	cc := NewTplConversionCtx()
	cc.nonce = "123"

	tests := []struct {
		name         string
		getToken     func() tplTypedToken
		panics       bool
		unwrappedVal any
	}{
		{
			name:         "IntAsInt",
			getToken:     func() tplTypedToken { return cc.AsInt(123) },
			unwrappedVal: int64(123),
		},
		{
			name:         "StrAsInt",
			getToken:     func() tplTypedToken { return cc.AsInt("123") },
			unwrappedVal: int64(123),
		},
		{
			name:     "NilAsInt",
			getToken: func() tplTypedToken { return cc.AsInt(nil) },
			panics:   true,
		},
		{
			name:     "SliceAsInt",
			getToken: func() tplTypedToken { return cc.AsInt([]string{"a"}) },
			panics:   true,
		},
		{
			name:         "IntAsFloat",
			getToken:     func() tplTypedToken { return cc.AsFloat(123) },
			unwrappedVal: float64(123),
		},
		{
			name:         "FloatAsFloat",
			getToken:     func() tplTypedToken { return cc.AsFloat(123.123) },
			unwrappedVal: float64(123.123),
		},
		{
			name:         "StringAsFloat",
			getToken:     func() tplTypedToken { return cc.AsFloat("123.123") },
			unwrappedVal: float64(123.123),
		},
		{
			name:         "BoolAsBool",
			getToken:     func() tplTypedToken { return cc.AsBool(true) },
			unwrappedVal: true,
		},
		{
			name:         "StringAsBool",
			getToken:     func() tplTypedToken { return cc.AsBool("false") },
			unwrappedVal: false,
		},
		{
			name:         "NilAsBool",
			getToken:     func() tplTypedToken { return cc.AsBool(nil) },
			unwrappedVal: false,
		},
		{
			name:         "NilAsNullable",
			getToken:     func() tplTypedToken { return cc.AsNullable(nil) },
			unwrappedVal: nil,
		},
		{
			name:         "StringAsNullable",
			getToken:     func() tplTypedToken { return cc.AsNullable("") },
			unwrappedVal: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.panics {
				assert.Panics(t, func() { _ = tt.getToken() })
				return
			}

			token := tt.getToken()

			assert.Equal(t, tt.unwrappedVal, cc.Unwrap(token.String()), "unexpected Unwrap() value")

		})
	}
}
