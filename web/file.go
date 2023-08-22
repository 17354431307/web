package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type FileUploader struct {
	FileFiled string

	// 为什么要用户传?
	// 要考虑文件名冲突的问题
	// 所以很多时候, 目标文件名字, 都是随机的
	DstPathFunc func(header *multipart.FileHeader) string
}

func (u FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		// 上传文件的逻辑在这里

		// 第一步: 读到文件内容

		// 第二步: 计算出目标路径
		// 第三步: 保存文件
		// 第四步: 返回响应

		file, fileheader, err := ctx.Req.FormFile(u.FileFiled)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传失败" + err.Error())
			return
		}
		defer file.Close()

		// 我怎么知道目标路径
		// 这种做法就是, 将目标路径的计算逻辑, 交给用户
		dst := u.DstPathFunc(fileheader)

		// Todo: 可以尝试把 dst 上不存在的目录给建立起来, os.MakeDirAll
		// os.MakeDirAll 可以把路径上不存在的路径给建立起来, 即使目录已经存在
		//os.MkdirAll(path)

		// O_WRONLY 可读可写
		// O_TRUNC 如果文件存在, 清空数据
		// O_CREATE 创建一个新的
		dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
		defer dstFile.Close()

		_, err = io.CopyBuffer(dstFile, file, nil)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传失败")
			return
		}

		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("上传成功")
	}
}
