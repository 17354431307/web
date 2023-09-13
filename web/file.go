package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	lru "github.com/hashicorp/golang-lru/v2"
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

type FileDownloader struct {
	Dir string
}

func (d FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		// 用的是 xxx?file=xxx
		value, err := ctx.QueryValue("file")
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("找不到目标文件")
			return
		}
		value = filepath.Clean(value)
		dst := filepath.Join(d.Dir, value)

		// 防止相对路径引起攻击者下载了你的系统文件
		//abs, err := filepath.Abs(dst)
		//if strings.Contains(dst, d.Dir) {
		//}

		fn := filepath.Base(dst)

		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")                     // 这两个选项是控制缓存的选项, 当前是不用缓存
		header.Set("Cache-Control", "must-revalidate") // 这两个选项是控制缓存的选项, 当前是不用缓存
		header.Set("Pragma", "public")

		http.ServeFile(ctx.Resp, ctx.Req, dst)
	}
}

type StaticResourceHandler struct {
	dir               string
	extContentTypeMap map[string]string
	cache             *lru.Cache[string, []byte]
	maxSize           int // 大文件不缓存
}

type StaticResourceHandlerOption func(handler *StaticResourceHandler)

func StaticWithMaxFileSize(maxSize int) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.maxSize = maxSize
	}
}
func StaticWithCache(c *lru.Cache[string, []byte]) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.cache = c
	}
}
func StaticWithExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		for ext, contentType := range extMap {
			handler.extContentTypeMap[ext] = contentType
		}
	}
}

// 两个层面上
// 1. 大文件不缓存
// 2. 控制住了缓存文件的数量
// 所以, 最多消耗多少内存? size(cache) * maxSize

func NewStaticResourceHandler(dir string, opts ...StaticResourceHandlerOption) (*StaticResourceHandler, error) {
	// 总共缓存 key-value
	c, err := lru.New[string, []byte](1000)
	if err != nil {
		return nil, err
	}

	res := &StaticResourceHandler{
		dir:   dir,
		cache: c,
		// 10 M 文件大小, 超过这个值, 就不会缓存
		maxSize: 1 << 20 * 10,
		extContentTypeMap: map[string]string{
			"jpg":  "image/jpeg",
			"jpe":  "image/jpeg",
			"jpeg": "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func (s *StaticResourceHandler) Handle(ctx *Context) {
	// 无缓存
	// 1. 拿到目标文件名
	// 2. 定位到目标文件, 并且读出来
	// 3. 返回给前端
	file, err := ctx.PathValue("file")
	if err != nil {
		ctx.RespStatusCode = http.StatusBadRequest
		ctx.RespData = []byte("请求路径不对")
		return
	}

	header := ctx.Resp.Header()
	// 可能的有文本文件, 图片, 多媒体(音频, 视频)
	ext := filepath.Ext(file)
	if data, ok := s.cache.Get(file); ok {

		header.Set("Content-Type", s.extContentTypeMap[ext[1:]])
		header.Set("Content-Language", strconv.Itoa(len(data)))
		ctx.RespData = data
		ctx.RespStatusCode = http.StatusOK
		return
	}

	dst := filepath.Join(s.dir, file)
	data, err := os.ReadFile(dst)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		ctx.RespData = []byte("服务器错误")
		return
	}

	// 大文件不缓存
	if len(data) <= s.maxSize {
		s.cache.Add(file, data)
	}

	header.Set("Content-Type", s.extContentTypeMap[ext[1:]])
	header.Set("Content-Language", strconv.Itoa(len(data)))
	ctx.RespData = data
	ctx.RespStatusCode = http.StatusOK

}
