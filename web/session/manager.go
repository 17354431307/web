package session

import (
	"github.com/Moty1999/web/web"
	"github.com/google/uuid"
)

type Manager struct {
	Propagator
	Store
	CtxSessKey string
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}
	val, ok := ctx.UserValues[m.CtxSessKey]
	if ok {
		return val.(Session), nil
	}

	sessId, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}

	sess, err := m.Get(ctx.Req.Context(), sessId)
	if err != nil {
		return nil, err
	}

	ctx.UserValues[m.CtxSessKey] = sess
	return sess, nil
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {

	id := uuid.New().String()
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}

	// 注入 HTTP 响应里面
	err = m.Inject(id, ctx.Resp)
	return sess, err
}

func (m *Manager) RemoveSession(ctx *web.Context) error {

	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}

	return m.Propagator.Remove(ctx.Resp)
}

func (m *Manager) RefreshSession(ctx *web.Context) error {

	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	return m.Refresh(ctx.Req.Context(), sess.ID())
}
