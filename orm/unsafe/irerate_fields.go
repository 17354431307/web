package unsafe

import "reflect"

/*
	Go按照字长对齐，因为 Go 本身每一次访问内存都是按照字长的倍数来访问的。
	+ 在 32 位字长机器上，就是按照 4 个字节对齐
	+ 在 64 位字长机器上，就是按照 8 个字节对齐


	但是有一个特别的点，
	比如 int32 占 4个字节，没有占满 8 个字节，如果下一个字段刚好和当前 int32 加起来占了 8 个字节，那莫下一个字段不会另起一行，会接着在 int32 后面排
	但是如果下个字段加当前 int32 超过了 8 个字节，Go 就会让下一个字段另起一行排放。
*/

func PrintFieldOffset(entity any) {
	typ := reflect.TypeOf(entity)
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		println(field.Offset)
	}
}
