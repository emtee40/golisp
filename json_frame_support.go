// Copyright 2014 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file implements Json<->Lisp conversions using frames.

package golisp

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func JsonToLispWithFrames(json interface{}) (result *Data) {
	mapValue, ok := json.(map[string]interface{})
	if ok {
		var m = make(FrameMap, len(mapValue))
		for key, val := range mapValue {
			value := JsonToLispWithFrames(val)
			m[fmt.Sprintf("%s:", key)] = value
		}
		return FrameWithValue(&m)
	}

	arrayValue, ok := json.([]interface{})
	if ok {
		var ary *Data
		for _, val := range arrayValue {
			value := JsonToLispWithFrames(val)
			ary = Cons(value, ary)
		}
		return Reverse(ary)
	}

	intValue, ok := json.(float64)
	if ok {
		return IntegerWithValue(int64(intValue))
	}

	strValue, ok := json.(string)
	if ok {
		return StringWithValue(strValue)
	}

	boolValue, ok := json.(bool)
	if ok {
		return BooleanWithValue(boolValue)
	}

	return
}

func JsonStringToLispWithFrames(jsonData string) (result *Data) {
	b := []byte(jsonData)
	var data interface{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Badly formed json: '%s'", jsonData)))
	}
	return JsonToLispWithFrames(data)
}

func LispWithFramesToJson(d *Data) (result interface{}) {
	if d == nil {
		return ""
	}

	if IntegerP(d) {
		return IntegerValue(d)
	}

	if StringP(d) {
		return StringValue(d)
	}

	if BooleanP(d) {
		return BooleanValue(d)
	}

	if PairP(d) {
		ary := make([]interface{}, 0, Length(d))
		for c := d; NotNilP(c); c = Cdr(c) {
			ary = append(ary, LispToJson(Car(c)))
		}
		return ary
	}

	if FrameP(d) {
		dict := make(map[string]interface{}, Length(d))
		for k, v := range *d.Frame {
			dict[strings.TrimRight(k, ":")] = LispWithFramesToJson(v)
		}
		return dict
	}

	return ""
}

func LispWithFramesToJsonString(d *Data) (result string) {
	temp := LispWithFramesToJson(d)
	j, err := json.Marshal(temp)
	if err == nil {
		return string(j)
	} else {
		return ""
	}
}
