package restruct

import (
	"fmt"
	"reflect"

	"github.com/alecthomas/repr"
)

// func init() {
// 	log.SetFlags(0)
// }

func genericTypeName[T any]() string {
	var zero *T
	return reflect.TypeOf(zero).Elem().String()
}

type RecursiveFunc[Target, Source any] func(cnv Converter, src Source) (Target, error)

func (RecursiveFunc[Target, Source]) Info() (string, string) {
	return genericTypeName[Target](), genericTypeName[Source]()
}

func (rf RecursiveFunc[Target, Source]) Convert(cnv Converter, source reflect.Value) (reflect.Value, error) {
	target, err := rf(cnv, source.Interface().(Source))
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(target), nil
}

type Func[Target, Source any] func(Source) (Target, error)

func (Func[Target, Source]) Info() (string, string) {
	return genericTypeName[Target](), genericTypeName[Source]()
}

func (nc Func[Target, Source]) Convert(cnv Converter, source reflect.Value) (reflect.Value, error) {
	target, err := nc(source.Interface().(Source))
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(target), nil
}

type MustFunc[Target, Source any] func(Source) Target

func (MustFunc[Target, Source]) Info() (string, string) {
	return genericTypeName[Target](), genericTypeName[Source]()
}

func (nc MustFunc[Target, Source]) Convert(cnv Converter, source reflect.Value) (reflect.Value, error) {
	target := nc(source.Interface().(Source))
	return reflect.ValueOf(target), nil
}

// StructFromStruct is a type that maps fields from a source struct to a target struct. The map keys are the target struct fields and the values are the source struct fields.
type StructFromStruct[Target, Source any] map[string]string

func (StructFromStruct[Target, Source]) Info() (string, string) {
	return genericTypeName[Target](), genericTypeName[Source]()
}

func (fm StructFromStruct[Target, Source]) Convert(cnv Converter, source reflect.Value) (reflect.Value, error) {
	var targetZero Target

	targetType := reflect.TypeOf(targetZero)

	target := reflect.New(targetType).Elem()

	if targetType.Kind() == reflect.Ptr {
		target = reflect.New(targetType.Elem()).Elem()
	}

	if source.Kind() == reflect.Interface || source.Kind() == reflect.Ptr {
		source = source.Elem()
	}

	for targetKey, sourceKey := range fm {
		targetField := target.FieldByName(targetKey)
		sourceField := source.FieldByName(sourceKey)

		if !sourceField.IsValid() || !targetField.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %s not found in source or target", sourceKey)
		}

		if !sourceField.Type().AssignableTo(targetField.Type()) {
			converted, err := cnv.Convert(sourceField, targetField.Type())
			if err != nil {
				return reflect.Value{}, err
			}

			sourceField = converted
		}

		targetField.Set(sourceField)
	}

	if targetType.Kind() == reflect.Ptr {
		target = target.Addr()
	}

	return target, nil
}

type ValueConverter interface {
	Info() (string, string)
	Convert(cnv Converter, source reflect.Value) (reflect.Value, error)
}

type Option interface{}

type Converter struct {
	conversions map[string]ValueConverter
}

func ConvertWith[T any](cnv Converter, v any) (T, error) {
	var targetZero T

	sourceValue := reflect.ValueOf(v)

	targetType := reflect.TypeOf(targetZero)

	targetValue, err := cnv.Convert(sourceValue, targetType)
	if err != nil {
		return targetZero, err
	}

	return targetValue.Interface().(T), nil
}

func (c Converter) Convert(source reflect.Value, targetType reflect.Type) (v reflect.Value, err error) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("conversion failed: %v", r)
	// 	}
	// }()

	for source.Kind() == reflect.Ptr || source.Kind() == reflect.Interface {
		source = source.Elem()
	}

	// check if there is a "specific" stored conversion
	specificConv, ok := c.conversions[targetType.String()+"<-"+source.Type().String()]
	if ok {
		return specificConv.Convert(c, source)
	}

	// then try to find a generic conversion
	genericConv, ok := c.conversions[targetType.String()]
	if ok {
		return genericConv.Convert(c, source)
	}

	// check if the source is already convertible to the target
	if source.Type().ConvertibleTo(targetType) {
		return source.Convert(targetType), nil
	}

	repr.Println(c.conversions)
	return reflect.Value{}, fmt.Errorf("no conversion found for %s", targetType.String())
}

func Convert[T any](v any, options ...Option) (T, error) {
	var targetZero T

	cnv := Converter{
		conversions: map[string]ValueConverter{},
	}

	for _, opt := range options {
		switch opt := opt.(type) {
		case ValueConverter:
			targetType, sourceType := opt.Info()

			// HACK: store a generic conversion and a specific one to support interface conversions
			cnv.conversions[targetType] = opt
			cnv.conversions[targetType+"<-"+sourceType] = opt
		default:
			panic("Invalid option type")
		}
	}

	// fmt.Println("Conversions:")
	// for key, conv := range cnv.conversions {
	// 	fmt.Printf("%s : %#v\n", key, conv)
	// }
	// fmt.Println()

	sourceValue := reflect.ValueOf(v)

	targetType := reflect.TypeOf(targetZero)

	recover()

	targetValue, err := cnv.Convert(sourceValue, targetType)
	if err != nil {
		return targetZero, err
	}

	return targetValue.Interface().(T), nil
}
