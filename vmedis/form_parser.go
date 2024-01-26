package vmedis

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FormUnmarshaler is the interface that wraps the UnmarshalForm method.
type FormUnmarshaler interface {
	UnmarshalForm(selection *goquery.Selection) error
}

var (
	formUnmarshalerType = reflect.TypeOf((*FormUnmarshaler)(nil)).Elem()
)

// UnmarshalForm parses the default value of a form from the API response.
func UnmarshalForm(selection *goquery.Selection, v any) error {
	return unmarshalForm(selection, reflect.ValueOf(v))
}

func unmarshalForm(selection *goquery.Selection, vv reflect.Value) error {
	vt := vv.Type()
	if vt.Implements(formUnmarshalerType) {
		return vv.Interface().(FormUnmarshaler).UnmarshalForm(selection)
	}

	if reflect.PointerTo(vt).Implements(formUnmarshalerType) {
		return vv.Addr().Interface().(FormUnmarshaler).UnmarshalForm(selection)
	}

	if vv.Kind() == reflect.Ptr {
		return unmarshalForm(selection, vv.Elem())
	}

	if !vv.CanSet() {
		return nil
	}

	switch vv.Kind() {
	case reflect.Struct:
		for i := 0; i < vv.NumField(); i++ {
			f, ft := vv.Field(i), vv.Type().Field(i)
			if ft.Type.Kind() != reflect.Slice {
				dataTag, ok := ft.Tag.Lookup("form-name")
				if !ok {
					continue
				}

				var data *goquery.Selection
				if dataTag == "<self>" {
					data = selection
				} else {
					data = selection.Find(fmt.Sprintf("input[name='%s']", dataTag)).First()
				}

				if data.Length() > 0 {
					if err := unmarshalForm(data, f); err != nil {
						return fmt.Errorf("unmarshal field %s: %w", ft.Name, err)
					}
				}
			} else {
				dataTagTemplate, ok := ft.Tag.Lookup("form-name")
				if !ok {
					continue
				}

				if strings.Count(dataTagTemplate, "%d") != 1 {
					return fmt.Errorf("form-name tag must contain exactly one %%d for slice field %s", ft.Name)
				}

				var result []reflect.Value
				for i := 0; ; i++ {
					var dataTag string
					if i == 0 {
						dataTag = strings.ReplaceAll(dataTagTemplate, "%d", "")
					} else {
						dataTag = fmt.Sprintf(dataTagTemplate, i)
					}

					data := selection.Find(fmt.Sprintf("input[name='%s']", dataTag)).First()
					if data.Length() == 0 {
						break
					}

					elem := reflect.New(ft.Type.Elem()).Elem()
					if err := unmarshalForm(data, elem); err != nil {
						return fmt.Errorf("unmarshal field %s: %w", ft.Name, err)
					}

					result = append(result, elem)
				}

				slice := reflect.MakeSlice(ft.Type, len(result), len(result))
				for i, elem := range result {
					slice.Index(i).Set(elem)
				}

				f.Set(slice)
			}
		}

	case reflect.String:
		vv.SetString(selection.AttrOr("value", ""))
	}

	return nil
}
