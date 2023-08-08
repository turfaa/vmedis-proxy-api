package client

import (
	"fmt"
	"reflect"

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
func UnmarshalDataColumn(selection *goquery.Selection, v any) error {
	return unmarshalDataColumn(selection, reflect.ValueOf(v))
}

func unmarshalDataColumn(selection *goquery.Selection, vv reflect.Value) error {
	vt := vv.Type()
	if vt.Implements(dataColumnUnmarshalerType) {
		return vv.Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if reflect.PointerTo(vt).Implements(dataColumnUnmarshalerType) {
		return vv.Addr().Interface().(DataColumnUnmarshaler).UnmarshalDataColumn(selection)
	}

	if vv.Kind() == reflect.Ptr {
		return unmarshalDataColumn(selection, vv.Elem())
	}

	if !vv.CanSet() {
		return fmt.Errorf("cannot set value of type %s", vv.Type())
	}

	switch vv.Kind() {
	case reflect.Struct:
		for i := 0; i < vv.NumField(); i++ {
			f, ft := vv.Field(i), vv.Type().Field(i)
			dataTag, ok := ft.Tag.Lookup("data-column")
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
				if err := unmarshalDataColumn(data, f); err != nil {
					return fmt.Errorf("unmarshal field %s: %w", ft.Name, err)
				}
			}
		}

	case reflect.String:
		vv.SetString(selection.Text())
	}

	return nil
}
