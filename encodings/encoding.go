package encodings

import (
	"encoding/binary"
	"errors"
	"github.com/vprix/vncproxy/pkg/dbuffer"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

func ReadUint8(r io.Reader) (uint8, error) {
	var myUint uint8
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	var myUint uint16
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	var myUint uint32
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadBytes(count int, r io.Reader) ([]byte, error) {
	buff := dbuffer.GetByteBuffer()
	defer dbuffer.ReleaseByteBuffer(buff)
	buff.ChangeLen(count)

	lengthRead, err := io.ReadFull(r, buff.B)

	if lengthRead != count {
		return nil, errors.New("ReadBytes unable to read bytes")
	}

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func ReadPixel(c io.Reader, pf *rfb.PixelFormat) ([]byte, error) {
	px := make([]byte, int(pf.BPP/8))
	if err := binary.Read(c, pf.Order(), &px); err != nil {
		return nil, err
	}
	return px, nil
}
