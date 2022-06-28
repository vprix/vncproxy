package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ClientMessageHandler vnc握手已结束，进入消息交互阶段
// 启动两个协程处理后续消息逻辑
// 1. 协程1：通过ClientMessageCh通道获取消息，并把该消息写入到vnc服务端会话中。
// 2. 协程2：从vnc服务端会话中读取消息类型及消息内容，组装该消息，发消息发送到ServerMessageCh通道中，供其他功能消费
// 3. 发送编码格式消息SetEncodings到vnc服务端
// 4. 发送帧数据请求消息FramebufferUpdateRequest到vnc服务端
type ClientMessageHandler struct{}

func (*ClientMessageHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debug("[Proxy客户端->VNC服务端]: vnc握手已结束，进入消息交互阶段[ClientMessageHandler]")
	}
	cfg := session.Options()
	var err error

	// proxy客户端支持的消息类型
	serverMessages := make(map[rfb.MessageType]rfb.Message)
	for _, m := range session.Messages() {
		serverMessages[m.Type()] = m
	}

	// 通过ClientMessageCh通道获取消息，并把该消息写入到vnc服务端会话中。
	go func() {
		for {
			select {
			case msg := <-cfg.Output:
				if logger.IsDebug() {
					logger.Debugf("[Proxy客户端->VNC服务端] 消息类型:%s,消息内容:%s", rfb.ClientMessageType(msg.Type()), msg.String())
				}
				if err = msg.Write(session); err != nil {
					cfg.ErrorCh <- err
					return
				}
			}
		}
	}()

	// 从vnc服务端会话中读取消息类型及消息内容，组装该消息，发消息发送到ServerMessageCh通道中，供其他功能消费
	go func() {
		for {
			select {
			default:
				// 从会话中读取消息类型
				var messageType rfb.MessageType
				if err = binary.Read(session, binary.BigEndian, &messageType); err != nil {
					cfg.ErrorCh <- err
					return
				}
				if logger.IsDebug() {
					logger.Debugf("[VNC服务端->Proxy客户端] 消息类型:%s", rfb.ClientMessageType(messageType))
				}
				// 判断proxy客户端是否支持该消息
				msg, ok := serverMessages[messageType]
				if !ok {
					err = fmt.Errorf("未知的消息类型: %v", messageType)
					cfg.ErrorCh <- err
					return
				}
				// 读取消息内容
				parsedMsg, err := msg.Read(session)
				if err != nil {
					cfg.ErrorCh <- err
					return
				}
				if logger.IsDebug() {
					logger.Debugf("[VNC服务端->Proxy客户端] 消息类型:%s,消息内容:%s", rfb.ClientMessageType(parsedMsg.Type()), parsedMsg)
				}
				cfg.Input <- parsedMsg
			}
		}
	}()
	return nil
}
