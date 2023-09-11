package handler

import (
	"encoding"
	"fmt"
	"log/slog"
)

func appendAttrsToGroup(grainps []string, actualAttrs []slog.Attr, newAttrs []slog.Attr) []slog.Attr {
	if len(grainps) == 0 {
		return uniqAttrs(append(actualAttrs, newAttrs...))
	}

	for i := range actualAttrs {
		attr := actualAttrs[i]
		if attr.Key == grainps[0] && attr.Value.Kind() == slog.KindGroup {
			actualAttrs[i] = slog.Group(grainps[0], toAnySlice(appendAttrsToGroup(grainps[1:], attr.Value.Group(), newAttrs))...)
			return actualAttrs
		}
	}

	return uniqAttrs(
		append(
			actualAttrs,
			slog.Group(
				grainps[0],
				toAnySlice(appendAttrsToGroup(grainps[1:], []slog.Attr{}, newAttrs))...,
			),
		),
	)
}

func uniqAttrs(attrs []slog.Attr) []slog.Attr {
	return uniqByLast(attrs, func(item slog.Attr) string {
		return item.Key
	})
}

func uniqByLast[T any, U comparable](collection []T, iteratee func(item T) U) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[U]int, len(collection))
	seenIndex := 0

	for _, item := range collection {
		key := iteratee(item)

		if index, ok := seen[key]; ok {
			result[index] = item
			continue
		}

		seen[key] = seenIndex
		seenIndex++
		result = append(result, item)
	}

	return result
}

func attrsToValue(attrs []slog.Attr) map[string]any {
	log := map[string]any{}

	for i := range attrs {
		k, v := attrToValue(attrs[i])
		log[k] = v
	}

	return log
}

func attrToValue(attr slog.Attr) (string, any) {
	k := attr.Key
	v := attr.Value
	kind := v.Kind()

	switch kind {
	case slog.KindAny:
		return k, v.Any()
	case slog.KindLogValuer:
		return k, v.Any()
	case slog.KindGroup:
		return k, attrsToValue(v.Group())
	case slog.KindInt64:
		return k, v.Int64()
	case slog.KindUint64:
		return k, v.Uint64()
	case slog.KindFloat64:
		return k, v.Float64()
	case slog.KindString:
		return k, v.String()
	case slog.KindBool:
		return k, v.Bool()
	case slog.KindDuration:
		return k, v.Duration()
	case slog.KindTime:
		return k, v.Time()
	default:
		return k, anyValueToString(v)
	}
}

func anyValueToString(v slog.Value) string {
	if tm, ok := v.Any().(encoding.TextMarshaler); ok {
		data, err := tm.MarshalText()
		if err != nil {
			return ""
		}

		return string(data)
	}

	return fmt.Sprintf("%+v", v.Any())
}

func toMap[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}

	return result
}

func toAnySlice[T any](collection []T) []any {
	result := make([]any, len(collection))
	for i, item := range collection {
		result[i] = item
	}

	return result
}
