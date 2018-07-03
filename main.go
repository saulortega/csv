package csv

import (
	gocsv "encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

type Comma rune

type UseCRLF bool

type Struct struct {
	Indexes    []int
	Header     map[int]string
	Types      map[int]string
	OmitHeader bool
}

type Whitelist []string

type Blacklist []string

type Header []string

func WriteTo(w io.Writer, obj interface{}, list ...interface{}) error {
	data, err := Format(obj, list...)
	if err != nil {
		return err
	}

	wr := gocsv.NewWriter(w)

	for _, l := range list {
		switch l.(type) {
		case Comma:
			wr.Comma = rune(l.(Comma))
		case UseCRLF:
			wr.UseCRLF = bool(l.(UseCRLF))
		}
	}

	return wr.WriteAll(data)
}

func Format(obj interface{}, list ...interface{}) ([][]string, error) {
	var data = [][]string{}

	value := reflect.ValueOf(obj)

	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return data, errors.New("array or slice required, " + value.Kind().String() + " received.")
	}

	if value.Len() == 0 {
		return data, nil
	}

	strct, err := dataStruct(value.Index(0), 0, list...)
	if err != nil {
		return data, err
	}

	if !strct.OmitHeader {
		row := []string{}
		for _, i := range strct.Indexes {
			row = append(row, strct.Header[i])
		}
		data = append(data, row)
	}

	if len(strct.Indexes) == 0 {
		return data, nil //Effects of Whitelist or Blacklist
	}

	for i := 0; i < value.Len(); i++ {
		row := formatStruct(value.Index(i), strct.Indexes, strct.Types)
		data = append(data, row)
	}

	return data, nil
}

func formatStruct(value reflect.Value, indexes []int, types map[int]string) []string {
	P := formatRow(value, types, 0)
	var row = []string{}
	for _, idx := range indexes {
		row = append(row, P[idx])
	}
	return row
}

func formatRow(value reflect.Value, types map[int]string, index int) map[int]string {
	var M = make(map[int]string)

	for i := 0; i < value.NumField(); i++ {
		if value.Type().Field(i).Anonymous && value.Type().Kind() == reflect.Struct {
			m := formatRow(value.Field(i), types, index)

			for ih, n := range m {
				M[ih] = n
			}

			index += value.Field(i).NumField()
			continue
		}

		index++
		switch types[index-1] {
		case "time.Time":
			M[index-1] = value.Field(i).Interface().(time.Time).Format("2006-01-02 15:04:05")
		case "null.Time":
			M[index-1] = value.Field(i).Field(0).Interface().(time.Time).Format("2006-01-02 15:04:05")
		case "null.Float64", "null.Float32", "null.Int", "null.Int8", "null.Int16", "null.Int32", "null.Int64", "null.Uint", "null.Uint8", "null.Uint16", "null.Uint32", "null.Uint64", "null.Bool":
			M[index-1] = fmt.Sprintf("%v", value.Field(i).Field(0))
		default:
			M[index-1] = fmt.Sprintf("%v", value.Field(i))
		}
		//Los nulos deberían verificarse si son nulos...
	}

	return M
}

func dataStruct(value reflect.Value, index int, list ...interface{}) (Struct, error) {
	var S = Struct{
		Header: make(map[int]string),
		Types:  make(map[int]string),
	}

	var wl Whitelist
	var bl Blacklist
	var hd Header
	if len(list) > 0 {
		for i := 0; i < len(list); i++ {
			switch list[i].(type) {
			case Whitelist:
				wl = list[i].(Whitelist)
			case Blacklist:
				bl = list[i].(Blacklist)
			case Header:
				hd = list[i].(Header)
				//default:
				//return S, errors.New("parameter invalid, only Whitelist, Blacklist and Header are accepted")
			}
		}
	}

	if value.Kind() != reflect.Struct {
		return S, errors.New("Struct required, «" + value.Kind().String() + "» received")
	}

	for i := 0; i < value.NumField(); i++ {
		if value.Type().Field(i).Anonymous && value.Type().Kind() == reflect.Struct {
			ds, err := dataStruct(value.Field(i), index, list...)
			if err != nil {
				return S, err
			}

			S.Indexes = append(S.Indexes, ds.Indexes...)
			for ih, n := range ds.Header {
				S.Header[ih] = n
			}
			for it, n := range ds.Types {
				S.Types[it] = n
			}

			index += value.Field(i).NumField()
			continue
		}
		index++

		name := strings.TrimSpace(value.Type().Field(i).Tag.Get("csv"))
		if name == "-" {
			continue
		}
		if name == "" {
			name = value.Type().Field(i).Name
		}
		if len(wl) > 0 {
			var ewl bool
			for j := 0; j < len(wl); j++ {
				if name == strings.TrimSpace(wl[j]) {
					ewl = true
					break
				}
			}
			if !ewl {
				continue
			}
		}
		if len(wl) == 0 && len(bl) > 0 {
			var ebl bool
			for j := 0; j < len(bl); j++ {
				if name == strings.TrimSpace(bl[j]) {
					ebl = true
					break
				}
			}
			if ebl {
				continue
			}
		}

		S.Types[index-1] = value.Field(i).Type().String()
		S.Indexes = append(S.Indexes, index-1)
		S.Header[index-1] = name
	}

	if len(hd) > 0 {
		if len(hd) == 1 && hd[0] == "-" {
			S.OmitHeader = true
		} else {
			if len(hd) == len(S.Indexes) {
				var j = 0
				for _, i := range S.Indexes {
					S.Header[i] = hd[j]
					j++
				}
			} else {
				return S, errors.New(fmt.Sprintf("%[1]v values required for header, %[2]v received.", len(S.Indexes), len(hd)))
			}
		}
	}

	return S, nil
}
