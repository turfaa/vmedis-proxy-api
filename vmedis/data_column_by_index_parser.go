package vmedis

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// UnmarshalDataColumnByIndex parses the data column from the API response.
// The data column is identified by its index.
func UnmarshalDataColumnByIndex(tag string, selection *goquery.Selection, v any) error {
	return unmarshalDataColumnByIndex(tag, selection, reflect.ValueOf(v))
}

func unmarshalDataColumnByIndex(tag string, selection *goquery.Selection, vv reflect.Value) error {
	vt := vv.Type()
	if vt.Implements(dataColumnUnmarshalerType) {
		return vv.Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if reflect.PointerTo(vt).Implements(dataColumnUnmarshalerType) {
		return vv.Addr().Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if vv.Kind() == reflect.Ptr {
		return unmarshalDataColumnByIndex(tag, selection, vv.Elem())
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
				data = selection.Find(fmt.Sprintf("td:nth-child(%s)", dataTag)).First()
			}

			if data.Length() > 0 {
				if err := unmarshalDataColumnByIndex(tag, data, f); err != nil {
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
