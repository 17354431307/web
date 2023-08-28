package session

import (
	"context"
	"net/http"
)

// Store 管理 Session 本身
type Store interface {
	// session 对应的 ID 谁来指定?
	// 要不要在接口维度上设置超时间, 以及, 要不要让 Store 内部去生成ID, 都是可以自由决策
	Generate(ctx context.Context, id string) (Session, error)
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
	// 这种也是可以的
	//Refresh(ctx context.Context, id string) error
}

type Session interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any) error
	ID() string
}

type Propagator interface {
	Inject(id string, writer http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Remove(writer http.ResponseWriter) error
}
