package geo

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
)

// ParseEWKBLonLatHex parses an EWKB hex string like
// "0101000020E610000093020B60CAC0C13F27F6D03E56CA4940"
// and returns lon, lat (float64). Works for POINT only.
func ParseEWKBLonLatHex(s string) (lon, lat float64, err error) {
	buf, err := hex.DecodeString(s)
	if err != nil {
		return 0, 0, err
	}

	if len(buf) < 1+4+16 { // byte order + type + x/y
		return 0, 0, errors.New("ewkb: too short")
	}

	// 1) Byte order
	var ord binary.ByteOrder

	switch buf[0] {
	case 0:
		ord = binary.BigEndian
	case 1:
		ord = binary.LittleEndian
	default:
		return 0, 0, errors.New("ewkb: invalid byte order")
	}

	// 2) Type/flags (EWKB)
	typ := ord.Uint32(buf[1:5])

	const sridFlag uint32 = 0x20000000
	hasSRID := (typ & sridFlag) != 0

	off := 5

	// 3) Optional SRID (we ignore the value, but need to skip it)
	if hasSRID {
		if len(buf) < off+4 {
			return 0, 0, errors.New("ewkb: missing srid")
		}

		_ = ord.Uint32(buf[off : off+4]) // srid
		off += 4
	}

	// 4) X (lon), Y (lat) as float64
	if len(buf) < off+16 {
		return 0, 0, errors.New("ewkb: missing coordinates")
	}

	lon = math.Float64frombits(ord.Uint64(buf[off : off+8]))
	off += 8
	lat = math.Float64frombits(ord.Uint64(buf[off : off+8]))

	return lon, lat, nil
}
