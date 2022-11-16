package target

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// tplValueType are the supported value types for type conversion
type tplValueType string

const (
	tplValueTypeInt      tplValueType = "int"
	tplValueTypeFloat    tplValueType = "float"
	tplValueTypeBool     tplValueType = "bool"
	tplValueTypeNullable tplValueType = "nullable"
)

// tplTypedToken represents a string token which encapsulates an output value of a specific type
// A typed token has the format "<prefix>:<nonce>:<type>:<value>"
type tplTypedToken string

const tplTypedTokenDelimiter = ":"
const tplTypedTokenMaxTokens = 4

func (t tplTypedToken) String() string { return string(t) }

func (t tplTypedToken) getNthToken(n int) string {
	if n >= tplTypedTokenMaxTokens {
		panic(fmt.Sprintf("cannot get a token larger than max token amount %d, got %d", tplTypedTokenMaxTokens, n))
	}
	values := strings.SplitN(t.String(), tplTypedTokenDelimiter, tplTypedTokenMaxTokens)
	if len(values) == tplTypedTokenMaxTokens {
		return values[n]
	}
	return ""
}

func (t tplTypedToken) GetPrefix() string {
	return t.getNthToken(0)
}

func (t tplTypedToken) GetNonce() string {
	return t.getNthToken(1)
}

func (t tplTypedToken) GetType() tplValueType {
	return tplValueType(t.getNthToken(2))
}

func (t tplTypedToken) GetValue() string {
	return t.getNthToken(3)
}

func (t tplTypedToken) IsWrapped(conversionCtx *tplTypeConversionContext) bool {
	prefix := t.GetPrefix()
	nonce := t.GetNonce()
	valueType := t.GetType()

	if prefix == "" || nonce == "" || valueType == "" {
		return false
	}

	if prefix != conversionCtx.prefix {
		return false
	}

	if nonce != conversionCtx.nonce {
		panic(fmt.Sprintf("string %s is wrapped with an incorrect nonce, expected %s, got %s", t.String(), conversionCtx.nonce, nonce))
	}

	switch valueType {
	case tplValueTypeFloat, tplValueTypeInt, tplValueTypeBool, tplValueTypeNullable:
		return true
	default:
		panic(fmt.Sprintf("string %s is wrapped with an unknown type: %s", t.String(), valueType))
	}
}

func (t tplTypedToken) Unwrap(conversionCtx *tplTypeConversionContext) any {
	if !t.IsWrapped(conversionCtx) {
		return t.String()
	}

	var retVal any
	var err error

	value := t.GetValue()
	valType := t.GetType()

	switch valType {
	case tplValueTypeFloat:
		retVal, err = strconv.ParseFloat(value, 64)
	case tplValueTypeInt:
		retVal, err = strconv.ParseInt(value, 10, 64)
	case tplValueTypeBool:
		retVal = !(strings.TrimSpace(value) == "" || value == "0" || strings.ToLower(value) == "false")
	case tplValueTypeNullable:
		if value == "" || value == "nil" || value == "null" {
			retVal = nil
		} else {
			retVal = value
		}
	}

	if err != nil {
		panic(fmt.Sprintf("unable to unwrap token '%s', with type %v: %v", t.String(), valType, err))
	}

	return retVal
}

func (t tplTypedToken) Wrap(asType tplValueType, conversionCtx tplTypeConversionContext) tplTypedToken {
	wrappedStr := strings.Join([]string{conversionCtx.prefix, conversionCtx.nonce, string(asType), t.String()}, tplTypedTokenDelimiter)
	return tplTypedToken(wrappedStr)
}

// tplTypeConversionContext holds the context for set of specific template type conversion execution
// This object holds the prefix and nonce used to generate [tplTypeToken]
type tplTypeConversionContext struct {
	prefix string
	nonce  string
}

func NewTplConversionCtx() tplTypeConversionContext {
	return tplTypeConversionContext{
		prefix: "fleetYamlTplTypeConv",
		nonce:  fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
	}
}

// AddFuncs adds the type conversion template functions in this context to the given [template.Template]
func (t *tplTypeConversionContext) AddFuncs(tplFn *template.Template) {
	conversionFuncs := template.FuncMap{
		"asInt":      t.AsInt,
		"asFloat":    t.AsFloat,
		"asBool":     t.AsBool,
		"asNullable": t.AsNullable,
	}

	tplFn.Funcs(conversionFuncs)
}

func (cc *tplTypeConversionContext) wrapped(valueType tplValueType, input any) tplTypedToken {
	str, ok := convertToStringsDeep(input).(string)
	if !ok && valueType != tplValueTypeNullable && valueType != tplValueTypeBool {
		panic(fmt.Sprintf("cannot convert %v to %v", input, valueType))
	}

	token := tplTypedToken(str)
	return token.Wrap(valueType, *cc)
}

// AsInt is a template function that wraps the input value to be unwrapped as an [int64]
func (cc *tplTypeConversionContext) AsInt(input any) tplTypedToken {
	return cc.wrapped(tplValueTypeInt, input)
}

// AsFloat is a template function that wraps the input value to be unwrapped as a [float64]
func (cc *tplTypeConversionContext) AsFloat(input any) tplTypedToken {
	return cc.wrapped(tplValueTypeFloat, input)
}

// AsBool is a template function that wraps the input value to be unwrapped as a [bool]
func (cc *tplTypeConversionContext) AsBool(input any) tplTypedToken {
	return cc.wrapped(tplValueTypeBool, input)
}

// AsNullable is a template function that wraps the input value to be unwrapped as nil when empty
func (cc *tplTypeConversionContext) AsNullable(input any) tplTypedToken {
	return cc.wrapped(tplValueTypeNullable, input)
}

// Unwrap attempts to accept a typed token string and unwrap it's type
func (cc *tplTypeConversionContext) Unwrap(input string) any {
	token := tplTypedToken(input)
	return token.Unwrap(cc)
}

func convertToStringsDeep(context any) any {
	switch val := context.(type) {
	case int, int32, int64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case map[string]any:
		newMap := make(map[string]any)
		for key, val := range val {
			newMap[key] = convertToStringsDeep(val)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(val))
		for i, v := range val {
			newSlice[i] = convertToStringsDeep(v)
		}
		return newSlice
	default:
		return val
	}
}
