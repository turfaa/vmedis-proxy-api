package vmedis

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DataColumnUnmarshaler is the interface that wraps the UnmarshalDataColumn method.
type DataColumnUnmarshaler interface {
	UnmarshalDataColumn(selection *goquery.Selection) error
}

var (
	dataColumnUnmarshalerType = reflect.TypeOf((*DataColumnUnmarshaler)(nil)).Elem()
)

// UnmarshalDataColumn parses the data column from the API response.
func UnmarshalDataColumn(tag string, selection *goquery.Selection, v any) error {
	return unmarshalDataColumn(tag, selection, reflect.ValueOf(v))
}

func unmarshalDataColumn(tag string, selection *goquery.Selection, vv reflect.Value) error {
	vt := vv.Type()
	if vt.Implements(dataColumnUnmarshalerType) {
		return vv.Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if reflect.PointerTo(vt).Implements(dataColumnUnmarshalerType) {
		return vv.Addr().Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if vv.Kind() == reflect.Ptr {
		return unmarshalDataColumn(tag, selection, vv.Elem())
	}

	if !vv.CanSet() {
		return fmt.Errorf("cannot set value of type %s", vv.Type())
	}

	selectionText := strings.TrimSpace(selection.Text())

	switch vv.Kind() {
	case reflect.Struct:
		for i := 0; i < vv.NumField(); i++ {
			f, ft := vv.Field(i), vv.Type().Field(i)
			dataTag, ok := ft.Tag.Lookup(tag)
			if !ok {
				continue
			}

			var data *goquery.Selection
			if dataTag == "<self>" {
				data = selection
			} else {
				data = selection.Find(fmt.Sprintf("td[data-col-seq='%s']", dataTag)).First()
			}

			if data.Length() > 0 {
				if err := unmarshalDataColumn(tag, data, f); err != nil {
					return fmt.Errorf("unmarshal field %s: %w", ft.Name, err)
				}
			}
		}

	case reflect.String:
		vv.SetString(selectionText)

	case reflect.Float32, reflect.Float64:
		f, err := parseFloat(selectionText)
		if err != nil {
			return fmt.Errorf("parse float from string [%s]: %w", selectionText, err)
		}

		vv.SetFloat(f)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(selectionText, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int from string [%s]: %w", selectionText, err)
		}

		vv.SetInt(i)
	}

	return nil
}
