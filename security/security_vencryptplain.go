package security

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

type ClientAuthVeNCrypt02Plain struct {
	Username []byte
	Password []byte
}

func (*ClientAuthVeNCrypt02Plain) Type() rfb.SecurityType {
	return rfb.SecTypeVeNCrypt
}

func (*ClientAuthVeNCrypt02Plain) SubType() rfb.SecuritySubType {
	return rfb.SecSubTypeVeNCrypt02Plain
}

func (auth *ClientAuthVeNCrypt02Plain) Auth(session rfb.ISession) error {
	// 发送认证版本号
	if err := binary.Write(session, binary.BigEndian, []uint8{0, 2}); err != nil {
		return err
	}
	if err := session.Flush(); err != nil {
		return err
	}
	var (
		major, minor uint8
	)
	// 对比版本号
	if err := binary.Read(session, binary.BigEndian, &major); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &minor); err != nil {
		return err
	}
	res := uint8(1)
	if major == 0 && minor == 2 {
		res = uint8(0)
	}
	if err := binary.Write(session, binary.BigEndian, res); err != nil {
		return err
	}
	if err := session.Flush(); err != nil {
		return err
	}
	// 选择认证子类型,只支持 SecSubTypeVeNCrypt02Plain 用户名密码认证
	if err := binary.Write(session, binary.BigEndian, uint8(1)); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, auth.SubType()); err != nil {
		return err
	}
	if err := session.Flush(); err != nil {
		return err
	}
	var secType rfb.SecuritySubType
	if err := binary.Read(session, binary.BigEndian, &secType); err != nil {
		return err
	}
	// 客户端选择的认证类型服务端不支持
	if secType != auth.SubType() {
		if err := binary.Write(session, binary.BigEndian, uint8(1)); err != nil {
			return err
		}
		if err := session.Flush(); err != nil {
			return err
		}
		return fmt.Errorf("invalid sectype")
	}
	// 服务端未设置用户名密码认证数据
	if len(auth.Password) == 0 || len(auth.Username) == 0 {
		return fmt.Errorf("Security Handshake failed; no username and/or password provided for VeNCryptAuth. ")
	}
	var (
		uLength, pLength uint32
	)
	// 获取用户名和密码长度
	if err := binary.Read(session, binary.BigEndian, &uLength); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &pLength); err != nil {
		return err
	}

	// 获取用户名和密码内容
	username := make([]byte, uLength)
	password := make([]byte, pLength)
	if err := binary.Read(session, binary.BigEndian, &username); err != nil {
		return err
	}

	if err := binary.Read(session, binary.BigEndian, &password); err != nil {
		return err
	}
	// 对比用户名密码是否正确，如果不正确则报错
	if !bytes.Equal(auth.Username, username) || !bytes.Equal(auth.Password, password) {
		return fmt.Errorf("invalid username/password")
	}
	return nil
}
