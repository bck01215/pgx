package pgproto3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/jackc/pgio"
)

const (
	TextFormat   = 0
	BinaryFormat = 1
)

type FieldDescription struct {
	Name                 []byte
	TableOID             uint32
	TableAttributeNumber uint16
	DataTypeOID          uint32
	DataTypeSize         int16
	TypeModifier         int32
	Format               int16
}

type RowDescription struct {
	Fields []FieldDescription
}

func (*RowDescription) Backend() {}

func (dst *RowDescription) Decode(src []byte) error {

	if len(src) < 2 {
		return &invalidMessageFormatErr{messageType: "RowDescription"}
	}
	fieldCount := int(binary.BigEndian.Uint16(src))
	rp := 2

	dst.Fields = dst.Fields[0:0]

	for i := 0; i < fieldCount; i++ {
		var fd FieldDescription

		idx := bytes.IndexByte(src[rp:], 0)
		if idx < 0 {
			return &invalidMessageFormatErr{messageType: "RowDescription"}
		}
		fd.Name = src[rp : rp+idx]
		rp += idx + 1

		// Since buf.Next() doesn't return an error if we hit the end of the buffer
		// check Len ahead of time
		if len(src[rp:]) < 18 {
			return &invalidMessageFormatErr{messageType: "RowDescription"}
		}

		fd.TableOID = binary.BigEndian.Uint32(src[rp:])
		rp += 4
		fd.TableAttributeNumber = binary.BigEndian.Uint16(src[rp:])
		rp += 2
		fd.DataTypeOID = binary.BigEndian.Uint32(src[rp:])
		rp += 4
		fd.DataTypeSize = int16(binary.BigEndian.Uint16(src[rp:]))
		rp += 2
		fd.TypeModifier = int32(binary.BigEndian.Uint32(src[rp:]))
		rp += 4
		fd.Format = int16(binary.BigEndian.Uint16(src[rp:]))
		rp += 2

		dst.Fields = append(dst.Fields, fd)
	}

	return nil
}

func (src *RowDescription) Encode(dst []byte) []byte {
	dst = append(dst, 'T')
	sp := len(dst)
	dst = pgio.AppendInt32(dst, -1)

	dst = pgio.AppendUint16(dst, uint16(len(src.Fields)))
	for _, fd := range src.Fields {
		dst = append(dst, fd.Name...)
		dst = append(dst, 0)

		dst = pgio.AppendUint32(dst, fd.TableOID)
		dst = pgio.AppendUint16(dst, fd.TableAttributeNumber)
		dst = pgio.AppendUint32(dst, fd.DataTypeOID)
		dst = pgio.AppendInt16(dst, fd.DataTypeSize)
		dst = pgio.AppendInt32(dst, fd.TypeModifier)
		dst = pgio.AppendInt16(dst, fd.Format)
	}

	pgio.SetInt32(dst[sp:], int32(len(dst[sp:])))

	return dst
}

func (src *RowDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   string
		Fields []FieldDescription
	}{
		Type:   "RowDescription",
		Fields: src.Fields,
	})
}
