// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/pineapple/msg-demo/backend/inbox/internal/logic"
	"github.com/pineapple/msg-demo/backend/inbox/internal/pkg/authctx"
	"github.com/pineapple/msg-demo/backend/inbox/internal/svc"
	"github.com/pineapple/msg-demo/backend/inbox/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SendMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendMessageRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if userID, ok := authctx.UserIDFromCtx(r.Context()); ok {
			if req.Channel == "personal" || req.SenderId == 0 {
				req.SenderId = userID
			}
		}

		l := logic.NewSendMessageLogic(r.Context(), svcCtx)
		resp, err := l.SendMessage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func ListMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMessagesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if userID, ok := authctx.UserIDFromCtx(r.Context()); ok {
			req.UserId = userID
		}

		l := logic.NewListMessagesLogic(r.Context(), svcCtx)
		resp, err := l.ListMessages(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func MarkMessageReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MarkReadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if userID, ok := authctx.UserIDFromCtx(r.Context()); ok {
			req.UserId = userID
		}

		l := logic.NewMarkMessageReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkMessageRead(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UnreadCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnreadCountRequest
		if userID, ok := authctx.UserIDFromCtx(r.Context()); ok {
			req.UserId = userID
		}

		l := logic.NewUnreadCountLogic(r.Context(), svcCtx)
		resp, err := l.UnreadCount(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
