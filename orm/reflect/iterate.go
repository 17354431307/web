package reflect

import (
	"reflect"
)

func IterateArrayOrSlice(entity any) ([]any, error) {
	value := reflect.ValueOf(entity)

	var res []any
	for i := 0; i < value.Len(); i++ {
		res = append(res, value.Index(i).Interface())
	}

	return res, nil
}

func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)

	resKeys := make([]any, 0, val.Len())
	resVals := make([]any, 0, val.Len())

	// 第一种遍历方式, 使用 keys
	//keys := val.MapKeys()
	//for _, key := range keys {
	//	v := val.MapIndex(key)
	//	resKeys = append(resKeys, key.Interface())
	//	resVals = append(resVals, v.Interface())
	//}

	// 第二种遍历方式, 使用迭代器
	itr := val.MapRange()
	for itr.Next() {
		resKeys = append(resKeys, itr.Key().Interface())
		resVals = append(resVals, itr.Value().Interface())
	}

	return resKeys, resVals, nil
}
