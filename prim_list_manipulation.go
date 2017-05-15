// Copyright 2014 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file contains the list manipulation primitive functions.

package golisp

import (
	"fmt"
)

func RegisterListManipulationPrimitives() {
	MakePrimitiveFunction("list", "*", ListImpl)
	MakePrimitiveFunction("circular-list", "*", CircularListImpl)
	MakeTypedPrimitiveFunction("make-list", "1|2", MakeListImpl, []uint32{IntegerType, AnyType})
	MakeTypedPrimitiveFunction("length", "1", ListLengthImpl, []uint32{ConsCellType})
	MakePrimitiveFunction("cons", "2", ConsImpl)
	MakePrimitiveFunction("cons*", ">=1", ConsStarImpl)
	MakeTypedPrimitiveFunction("reverse", "1", ReverseImpl, []uint32{ConsCellType})
	MakeTypedPrimitiveFunction("flatten", "1", FlattenImpl, []uint32{ConsCellType})
	MakeTypedPrimitiveFunction("flatten*", "1", RecursiveFlattenImpl, []uint32{ConsCellType})
	MakePrimitiveFunction("append", "*", AppendImpl)
	MakePrimitiveFunction("append!", "*", AppendBangImpl)
	MakeTypedPrimitiveFunction("copy", "1", CopyImpl, []uint32{ConsCellType})
	MakePrimitiveFunction("partition", "2|3", PartitionImpl)
	MakeTypedPrimitiveFunction("sublist", "3", SublistImpl, []uint32{ConsCellType, IntegerType, IntegerType})
	MakeTypedPrimitiveFunction("sort", "2", SortImpl, []uint32{ConsCellType, FunctionType | PrimitiveType})
}

func ListImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return args, nil
}

func CircularListImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	ConsValue(LastPair(args)).Cdr = args
	return args, nil
}

func MakeListImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	kVal := First(args)
	k := IntegerValue(kVal)
	var element *Data

	if k < 0 {
		err = ProcessError("make-list requires a non-negative integer as it's first argument.", env)
		return
	}

	if Length(args) == 1 {
		element = nil
	} else {
		element = Second(args)
	}

	var items []*Data
	items = make([]*Data, 0, k)
	for ; k > 0; k = k - 1 {
		items = append(items, element)
	}
	return ArrayToList(items), nil
}

func ListLengthImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	col := First(args)
	if !ListP(col) {
		err = ProcessError(fmt.Sprintf("length requires a list but was given %s.", String(col)), env)
		return
	}

	return IntegerWithValue(int64(Length(col))), nil
}

func ConsImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	car := First(args)
	cdr := Second(args)
	return Cons(car, cdr), nil
}

func ConsStarImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	ary := ToArray(args)
	l := Length(args) - 1
	return ArrayToListWithTail(ary[0:l], ary[l]), nil
}

func ReverseImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return Reverse(First(args)), nil
}

func FlattenImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return Flatten(First(args))
}

func RecursiveFlattenImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return RecursiveFlatten(First(args))
}

func AppendBangImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {

	for l := args; NotNilP(l); l = Cdr(l) {
		if !ListP(Car(l)) && NotNilP(Cdr(l)) {
			err = ProcessError(fmt.Sprintf("append! requires lists for all non-final arguments but was given %s.", String(Car(l))), env)
			return
		}
	}

	var prev *Data = nil
	for cell := args; NotNilP(cell); cell = Cdr(cell) {
		if prev != nil {
			ConsValue(LastPair(prev)).Cdr = Car(cell)
		}
		prev = Car(cell)
	}

	return First(args), nil
}

func AppendImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	// No args -> empty list
	if Length(args) == 0 {
		return
	}

	// step through args, accumulating elements
	var items []*Data = make([]*Data, 0, 10)
	var item *Data
	for cell := args; NotNilP(cell); cell = Cdr(cell) {
		lastArg := NilP(Cdr(cell))
		item = Car(cell)
		if !ListP(item) && !lastArg { // items other than the last NUST be lists
			err = ProcessError(fmt.Sprintf("append requires lists for all non-final arguments but was given %s.", String(item)), env)
			return
		}

		if lastArg {
			result = ArrayToListWithTail(items, item)
			return
		} else {
			for innerCell := item; NotNilP(innerCell); innerCell = Cdr(innerCell) {
				items = append(items, Car(innerCell))
			}
		}
	}
	result = ArrayToList(items)
	return
}

func CopyImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return Copy(First(args)), nil
}

func partitionBySize(size int64, step int64, l *Data, env *SymbolTableFrame) (result *Data, err error) {
	elements := ToArray(l)

	if size < 1 || size > int64(len(elements)) {
		err = ProcessError("partition requires a clump size that fits in the list.", env)
		return
	}
	if step < 1 {
		err = ProcessError("partition requires a positive step size.", env)
		return
	}

	pieces := make([]*Data, 0, 5)
	var start int64 = 0
	var end int64 = size
	for end <= int64(len(elements)) {
		pieces = append(pieces, ArrayToList(elements[start:end]))
		start = start + step
		end = start + size
	}

	return ArrayToList(pieces), nil
}

func partitionByPredicate(determiner *Data, l *Data, env *SymbolTableFrame) (result *Data, err error) {
	pieces := make([]*Data, 0, 5)
	falseSection := make([]*Data, 0, 5)
	trueSection := make([]*Data, 0, 5)
	var predicateResult *Data
	for c := l; NotNilP(c); c = Cdr(c) {
		predicateResult, err = ApplyWithoutEval(determiner, InternalMakeList(Car(c)), env)
		if err != nil {
			return
		}
		if BooleanValue(predicateResult) {
			trueSection = append(trueSection, Car(c))
		} else {
			falseSection = append(falseSection, Car(c))
		}
	}

	pieces = append(pieces, ArrayToList(trueSection))
	pieces = append(pieces, ArrayToList(falseSection))

	return ArrayToList(pieces), nil
}

func PartitionImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	determiner := First(args)
	if !IntegerP(determiner) && !FunctionOrPrimitiveP(determiner) {
		err = ProcessError("partition requires an integer or function as it's first argument.", env)
		return
	}

	var l *Data
	var step int64

	if IntegerP(determiner) {
		if IntegerP(Second(args)) {
			step = IntegerValue(Second(args))
			l = Third(args)
		} else {
			step = IntegerValue(determiner)
			l = Second(args)
		}
	} else {
		l = Second(args)
	}

	if !ListP(l) {
		err = ProcessError("partition requires a list as it's final argument.", env)
		return
	}

	if IntegerP(determiner) {
		return partitionBySize(IntegerValue(determiner), step, l, env)
	} else {
		return partitionByPredicate(determiner, l, env)
	}
}

func SublistImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	l := First(args)
	if !ListP(l) {
		err = ProcessError("sublist requires a list as it's first argument.", env)
		return
	}

	n := Second(args)
	first := int(IntegerValue(n))

	if first < 0 {
		err = ProcessError("sublist requires positive indecies.", env)
		return
	}

	n = Third(args)
	last := int(IntegerValue(n))

	if last < 0 {
		err = ProcessError("sublist requires positive indecies.", env)
		return
	}

	if first >= last {
		return
	}

	var cell *Data
	var i int
	for i, cell = 0, l; i < first && NotNilP(cell); i, cell = i+1, Cdr(cell) {
	}

	var items []*Data = make([]*Data, 0, Length(args))
	for ; i < last && NotNilP(cell); i, cell = i+1, Cdr(cell) {
		items = append(items, Car(cell))
	}
	result = ArrayToList(items)
	return
}

func mergeCompare(a *Data, b *Data, proc *Data, env *SymbolTableFrame) (result bool, err error) {
	if FunctionP(proc) {
		b, err := FunctionValue(proc).ApplyWithoutEval(InternalMakeList(a, b), env)
		if err == nil {
			result = BooleanValue(b)
		}
	} else {
		b, err := PrimitiveValue(proc).ApplyWithoutEval(InternalMakeList(a, b), env)
		if err == nil {
			result = BooleanValue(b)
		}
	}
	return
}

func merge(a []*Data, b []*Data, proc *Data, env *SymbolTableFrame) (result []*Data, err error) {
	var r = make([]*Data, len(a)+len(b))
	var i = 0
	var j = 0
	var comparison = false

	for i < len(a) && j < len(b) {
		comparison, err = mergeCompare(a[i], b[j], proc, env)
		if err != nil {
			return
		}
		if comparison {
			r[i+j] = a[i]
			i++
		} else {
			r[i+j] = b[j]
			j++
		}
	}

	for i < len(a) {
		r[i+j] = a[i]
		i++
	}
	for j < len(b) {
		r[i+j] = b[j]
		j++
	}

	return r, nil
}

func MergeSort(items []*Data, proc *Data, env *SymbolTableFrame) (result []*Data, err error) {
	if len(items) < 2 {
		return items, nil
	}

	var middle = len(items) / 2

	a, err := MergeSort(items[:middle], proc, env)
	if err != nil {
		return
	}

	b, err := MergeSort(items[middle:], proc, env)
	if err != nil {
		return
	}

	return merge(a, b, proc, env)
}

func SortImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	coll := First(args)
	if !ListP(coll) {
		err = ProcessError("sort requires a list as it's first argument.", env)
		return
	}
	proc := Second(args)
	sorted, err := MergeSort(ToArray(coll), proc, env)
	if err != nil {
		return
	}

	return ArrayToList(sorted), nil
}
