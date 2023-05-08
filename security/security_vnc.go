package security

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/vprix/vncproxy/rfb"
)

// ChallengeLen 随机认证串的长度
const ChallengeLen = 16

// ServerAuthVNC vnc服务端使用vnc auth认证方式
type ServerAuthVNC struct {
	Challenge []byte
	Password  []byte
	Crypted   []byte
}

var _ rfb.ISecurityHandler = new(ServerAuthVNC)
var _ rfb.ISecurityHandler = new(ClientAuthVNC)

func (*ServerAuthVNC) Type() rfb.SecurityType {
	return rfb.SecTypeVNC
}
func (*ServerAuthVNC) SubType() rfb.SecuritySubType {
	return rfb.SecSubTypeUnknown
}

// 写入随机字符串
func (that *ServerAuthVNC) writeChallenge(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Challenge); err != nil {
		return err
	}
	return session.Flush()
}

func (that *ServerAuthVNC) ReadChallenge(session rfb.ISession) error {
	var crypted [ChallengeLen]byte
	if err := binary.Read(session, binary.BigEndian, &crypted); err != nil {
		return err
	}
	that.Crypted = crypted[:]
	return nil
}

func (that *ServerAuthVNC) Auth(session rfb.ISession) error {

	if len(that.Challenge) != ChallengeLen {
		that.Challenge = grand.B(ChallengeLen)
	}

	if err := that.writeChallenge(session); err != nil {
		return err
	}
	if err := that.ReadChallenge(session); err != nil {
		return err
	}
	// 加密随机认证串，并把加密后的串与客户端穿过来的串进行对比，如果对比一致，则说明密码一致
	encrypted, err := AuthVNCEncode(that.Password, that.Challenge)
	if err != nil {
		return err
	}
	if !bytes.Equal(encrypted, that.Crypted) {
		return fmt.Errorf("密码错误")
	}
	return nil
}

// ClientAuthVNC vnc 客户端使用vnc auth认证方式
type ClientAuthVNC struct {
	Challenge []byte
	Password  []byte
}

func (*ClientAuthVNC) Type() rfb.SecurityType {
	return rfb.SecTypeVNC
}
func (*ClientAuthVNC) SubType() rfb.SecuritySubType {
	return rfb.SecSubTypeUnknown
}

func (that *ClientAuthVNC) Auth(session rfb.ISession) error {
	if len(that.Password) == 0 {
		return fmt.Errorf("安全认证失败，因为没有传入VNCAuth认证方式所用的密码")
	}
	var challenge [ChallengeLen]byte
	if err := binary.Read(session, binary.BigEndian, &challenge); err != nil {
		return err
	}
	// 使用密码对认证串加密
	encrypted, err := AuthVNCEncode(that.Password, challenge[:])
	if err != nil {
		return err
	}
	// 发送加密后的认证串
	if err = binary.Write(session, binary.BigEndian, encrypted); err != nil {
		return err
	}
	return session.Flush()
}

// AuthVNCEncode 加密随机认证串
func AuthVNCEncode(password []byte, challenge []byte) ([]byte, error) {
	if len(challenge) != ChallengeLen {
		return nil, fmt.Errorf("随机认证串的长度不正确，正确的应该是16字节")
	}
	// 截取密码的前八位，因为只有前八位才有用
	key := make([]byte, 8)
	copy(key, password)

	// 对密码的每个字节进行翻转
	for i := range key {
		key[i] = (key[i]&0x55)<<1 | (key[i]&0xAA)>>1 // Swap adjacent bits
		key[i] = (key[i]&0x33)<<2 | (key[i]&0xCC)>>2 // Swap adjacent pairs
		key[i] = (key[i]&0x0F)<<4 | (key[i]&0xF0)>>4 // Swap the 2 halves
	}

	// 使用密码对随即认证串进行加密
	cipher, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(challenge); i += cipher.BlockSize() {
		cipher.Encrypt(challenge[i:i+cipher.BlockSize()], challenge[i:i+cipher.BlockSize()])
	}

	return challenge, nil
}
