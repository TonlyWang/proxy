package handler

import (
	"fmt"
	"proxy/core/zinx/ziface"
)

// TestRouter Struct
type TestRouter struct {
	BaseHandler
}

func (h *TestRouter) PreHandle(request ziface.IRequest) error {
	//执行baseHandler.PreHandle
	h.BaseRouter.PreHandle(request)

	fmt.Println("per handle test...")

	//return fmt.Errorf("pre handler is error")
	return nil
}

func (h *TestRouter) Handle(request ziface.IRequest) error {
	fmt.Println("handle test...")

	//message
	fmt.Println("data+++++++msg.id: ", request.GetMsgID())
	//fmt.Println("data+++++++msg.data: ", string(request.GetData()))

	//request.GetConnection().SendMsg(201, []byte(`this is a test message from the server!!!`))
	//request.GetConnection().SendBuffMsg(201, []byte(`this is a test message from the server!!!`))

	tcpServer := request.GetConnection().GetTCPServer()
	fmt.Println("server-hc::::::", tcpServer.GetHeartBeat())

	conn := request.GetConnection()
	fmt.Println("conn-hc::::::", conn.GetHeartBeat())

	PushMessage(request, 101, 0, []byte("a-202311161930-b"))

	//return fmt.Errorf("handler is error")
	return nil
}

func (h *TestRouter) PostHandle(request ziface.IRequest) error {
	//执行baseHandler.PreHandle
	h.BaseRouter.PostHandle(request)

	fmt.Println("post handle test...")

	return nil
}
