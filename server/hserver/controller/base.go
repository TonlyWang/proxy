package controller

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"net/http"
	"proxy/core/server"
	"proxy/core/zinx/ziface"
	"proxy/library/logger"
	"proxy/server/global"
	"runtime"
	"strings"
)

func WrapHandle(handler func(*server.Request) *server.Response) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//debug.PrintStack()
				logger.Error(r.Context(), fmt.Sprintf(`[wrap handle] Error(1): %v`, err))
				logger.Error(r.Context(), fmt.Sprintf(`[wrap handle] User Id: %v, Client IP: %v`, r.Header.Get("user_id"), r.Header.Get("client_ip")))
				for i := 1; i < 20; i++ {
					if pc, file, line, ok := runtime.Caller(i); ok {
						funcName := runtime.FuncForPC(pc).Name() //获取函数名
						logger.Error(r.Context(), fmt.Sprintf(`[wrap handle] goroutine:%v, file:%s, function name:%s, line:%d`, pc, file, funcName, line))
					}
				}
				logger.Error(r.Context(), fmt.Sprintf(`[wrap handle] Error(2): %v`, err))
			}
		}()

		//params
		proxyId := r.Header.Get("proxy_id")
		serverId := r.Header.Get("server_id")
		//uin := r.Header.Get("uin")
		userId := r.Header.Get("user_id") //当前玩家ID
		clientIP := r.Header.Get("client_ip")
		playerId := r.Header.Get("player_id") //需要推送消息的玩家ID(包含当前玩家ID)
		traceId := r.Header.Get("trace_id")

		//参数判断
		if proxyId == "" || serverId == "" || userId == "" || clientIP == "" || playerId == "" || traceId == "" {
			writeResponseData(w, &server.Response{Code: 1000})
			return
		}

		//待推送消息的玩家ID
		playerIds := make([]int, 0, 20)
		for _, v := range strings.Split(playerId, ",") {
			if v != "" {
				playerIds = append(playerIds, cast.ToInt(v))
			}
		}
		if len(playerIds) == 0 {
			writeResponseData(w, &server.Response{Code: 1001})
			return
		}

		//获取当前玩家conn
		var conn ziface.IConnection
		uid := cast.ToUint64(userId)
		if uid > 0 {
			connection, err := global.GetTCPServer().GetConnMgr().GetConnByUserId(uid)
			if err != nil {
				writeResponseData(w, &server.Response{Code: 1005})
				return
			}
			if connection.GetProxyId() != cast.ToUint32(proxyId) {
				writeResponseData(w, &server.Response{Code: 1010})
				return
			}
			if connection.GetServerId() != cast.ToUint32(serverId) {
				writeResponseData(w, &server.Response{Code: 1012})
				return
			}
			if connection.GetUserId() != cast.ToUint64(userId) {
				writeResponseData(w, &server.Response{Code: 1015})
				return
			}
			conn = connection
		}

		//request
		request := &server.Request{
			ResponseWriter: w,
			Request:        r,
			PlayerID:       playerIds,
			UserID:         uid,
			Conn:           conn,
			Data:           make(map[string]any, 10),
		}
		request.SetTraceID(traceId)

		//handler
		resp := handler(request)
		if resp.Type == 1 {
			logger.Info(request, fmt.Sprintf(`[proxy id:%v, server id:%v, ip:%v, player id:%v, user id:%v, code:%v, data:%v, message:%v]`, proxyId, serverId, clientIP, playerId, userId, resp.Code, resp.Data, resp.Msg))
		} else if resp.Type == 2 {
			logger.Error(request, fmt.Sprintf(`[proxy id:%v, server id:%v, ip:%v, player id:%v, user id:%v, code:%v, data:%v, message:%v]`, proxyId, serverId, clientIP, playerId, userId, resp.Code, resp.Data, resp.Msg))
		}

		//result
		writeResponseData(w, resp)
	}
}

func writeResponseData(w http.ResponseWriter, params *server.Response) {
	dataByte, _ := json.Marshal(params)
	w.Header().Set("content-length", cast.ToString(len(dataByte)))
	w.Write(dataByte)
	w.(http.Flusher).Flush()
}

func writeResponseBytes(w http.ResponseWriter, data []byte) {
	w.Header().Set("content-length", cast.ToString(len(data)))
	w.Write(data)
	w.(http.Flusher).Flush()
}
