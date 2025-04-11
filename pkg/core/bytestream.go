package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
)

const (
	defaultInitialCapacity = 128
	maxStringLength        = 900_000

	EmptyString = "<empty>"
)

var (
	ErrStringTooLong        = errors.New("string is too long")
	ErrStringNegativeLength = errors.New("string has negative length")
)

type VInt int32

type LogicLong struct {
	F int32
	S int32
}

type DataRef struct {
	F int32
	S int32
}

type ScId DataRef

type ByteStream struct {
	buffer []byte

	offset    int
	bitOffset int
}

func NewByteStream(initialBuffer []byte) *ByteStream {
	var capacity = defaultInitialCapacity

	if initialBuffer != nil {
		if cap(initialBuffer) > capacity {
			capacity = cap(initialBuffer)
		}

		buf := make([]byte, len(initialBuffer), capacity)
		copy(buf, initialBuffer)

		return &ByteStream{
			buffer: buf,
		}
	}

	return &ByteStream{
		buffer: make([]byte, 0, capacity),
	}
}

func NewByteStreamWithCapacity(initialCapacity int) *ByteStream {
	if initialCapacity < 0 {
		initialCapacity = defaultInitialCapacity
	}

	return &ByteStream{
		buffer: make([]byte, 0, initialCapacity),
	}
}

// --- Buffer management --- //

func (b *ByteStream) Clear() {
	b.buffer = make([]byte, 0, cap(b.buffer))
	b.offset = 0
	b.bitOffset = 0
}

func (b *ByteStream) Close() {
	b.buffer = nil
	b.offset = 0
	b.bitOffset = 0
}

func (b *ByteStream) ensureReadCapacity(n int) error {
	if b.offset+n > len(b.buffer) {
		return io.EOF
	}

	return nil
}

// --- Read operations --- //

func (b *ByteStream) ReadBool() (bool, error) {
	if b.bitOffset == 0 {
		if err := b.ensureReadCapacity(1); err != nil {
			return false, fmt.Errorf("failed to read boolean: %w", err)
		}
	}

	byteIndex := b.offset

	if byteIndex >= len(b.buffer) {
		return false, fmt.Errorf("readBool internal logic error: index out of bounds")
	}

	current := b.buffer[byteIndex]
	value := (current >> b.bitOffset) & 1

	b.bitOffset++

	if b.bitOffset == 8 {
		b.bitOffset = 0
		b.offset++
	}

	return value != 0, nil
}

func (b *ByteStream) ReadInt() (int32, error) {
	b.bitOffset = 0

	if err := b.ensureReadCapacity(4); err != nil {
		return 0, fmt.Errorf("failed to read int: %w", err)
	}

	value := binary.BigEndian.Uint32(b.buffer[b.offset:])
	b.offset += 4

	return int32(value), nil
}

func (b *ByteStream) ReadString() (string, error) {
	length, err := b.ReadInt()

	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	if length == -1 {
		return "", nil
	}

	if length == 0 {
		return "", nil
	}

	if length < -1 {
		return "", fmt.Errorf("failed to read string: %w (length %d)", ErrStringNegativeLength, length)
	}

	if length > maxStringLength {
		return "", fmt.Errorf("failed to read string: %w (length %d)", ErrStringTooLong, length)
	}

	b.bitOffset = 0

	if err := b.ensureReadCapacity(int(length)); err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}

	strBytes := b.buffer[b.offset : b.offset+int(length)]
	b.offset += int(length)

	result := string(strBytes)
	return result, nil
}

func (b *ByteStream) ReadVInt() (VInt, error) {
	b.bitOffset = 0

	initialOffset := b.offset

	if initialOffset >= len(b.buffer) {
		return 0, io.EOF
	}

	firstByte := b.buffer[initialOffset]
	bytesRead := 1

	var result int32 = int32(firstByte & 0x3F)

	if (firstByte & 0x40) != 0 {
		if (firstByte & 0x80) != 0 {
			if initialOffset+bytesRead >= len(b.buffer) {
				return 0, io.EOF
			}

			byte2 := b.buffer[initialOffset+bytesRead]
			bytesRead++

			result = (result & ^int32(0x1FC0)) | (int32(byte2&0x7F) << 6)

			if (byte2 & 0x80) != 0 {
				if initialOffset+bytesRead >= len(b.buffer) {
					return 0, io.EOF
				}

				byte3 := b.buffer[initialOffset+bytesRead]
				bytesRead++

				result = (result & ^int32(0xFE000)) | (int32(byte3&0x7F) << 13)

				if (byte3 & 0x80) != 0 {
					if initialOffset+bytesRead >= len(b.buffer) {
						return 0, io.EOF
					}

					byte4 := b.buffer[initialOffset+bytesRead]
					bytesRead++

					result = (result & ^int32(0x7F00000)) | (int32(byte4&0x7F) << 20)

					if (byte4 & 0x80) != 0 {
						if initialOffset+bytesRead >= len(b.buffer) {
							return 0, io.EOF
						}

						byte5 := b.buffer[initialOffset+bytesRead]
						bytesRead++

						result = (result & 0x7FFFFFFF) | (int32(byte5&0x0F) << 27) | math.MinInt32
					} else {
						result |= int32(-134217728)
					}
				} else {
					result |= int32(-1048576)
				}
			} else {
				result |= int32(-8192)
			}
		} else {
			result |= int32(-64)
		}

	} else if (firstByte & 0x80) != 0 {
		if initialOffset+bytesRead >= len(b.buffer) {
			return 0, io.EOF
		}

		byte2 := b.buffer[initialOffset+bytesRead]
		bytesRead++

		result = (result & ^int32(0x1FC0)) | (int32(byte2&0x7F) << 6)

		if (byte2 & 0x80) != 0 {
			if initialOffset+bytesRead >= len(b.buffer) {
				return 0, io.EOF
			}

			byte3 := b.buffer[initialOffset+bytesRead]
			bytesRead++

			result = (result & ^int32(0xFE000)) | (int32(byte3&0x7F) << 13)

			if (byte3 & 0x80) != 0 {
				if initialOffset+bytesRead >= len(b.buffer) {
					return 0, io.EOF
				}

				byte4 := b.buffer[initialOffset+bytesRead]
				bytesRead++

				result = (result & ^int32(0x7F00000)) | (int32(byte4&0x7F) << 20)

				if (byte4 & 0x80) != 0 {
					if initialOffset+bytesRead >= len(b.buffer) {
						return 0, io.EOF
					}

					byte5 := b.buffer[initialOffset+bytesRead]
					bytesRead++

					result = (result & 0x7FFFFFFF) | (int32(byte5&0x0F) << 27)
				}
			}
		}
	}

	b.offset = initialOffset + bytesRead

	return VInt(result), nil
}

func (b *ByteStream) ReadDataRef() (DataRef, error) {
	classId, err := b.ReadVInt()

	if err != nil {
		return DataRef{}, fmt.Errorf("failed to read class id: %w", err)
	}

	if classId == 0 {
		return DataRef{F: 0, S: 0}, nil
	}

	instanceId, err := b.ReadVInt()

	if err != nil {
		return DataRef{}, fmt.Errorf("failed to read instance id (class id was %d): %w", classId, err)
	}

	return DataRef{
		F: int32(classId),
		S: int32(instanceId),
	}, nil
}

func (b *ByteStream) ReadScId() (ScId, error) {
	dataRef, err := b.ReadDataRef()

	if err != nil {
		return ScId{}, fmt.Errorf("failed to read underlying DataRef: %w", err)
	}

	return ScId(dataRef), nil
}

// --- Write operations --- //

func (b *ByteStream) writeBytesInternal(data []byte) {
	b.buffer = append(b.buffer, data...)
	b.offset = len(b.buffer)
}

func (b *ByteStream) WriteBool(value bool) {
	if b.bitOffset == 0 {
		b.buffer = append(b.buffer, 0)
		b.offset = len(b.buffer)
	}

	lastByteIndex := len(b.buffer) - 1

	if value {
		b.buffer[lastByteIndex] |= 1 << (b.bitOffset /* & 31 */)
	} else {
		b.buffer[lastByteIndex] &= ^(1 << b.bitOffset)
	}

	b.bitOffset = (b.bitOffset + 1) & 7
}

func (b *ByteStream) WriteInt(value int32) {
	b.bitOffset = 0

	var buf [4]byte

	binary.BigEndian.PutUint32(buf[:], uint32(value))
	b.writeBytesInternal(buf[:])
}

func (b *ByteStream) WriteString(value string) {
	b.bitOffset = 0

	if value == EmptyString {
		b.WriteInt(-1)
		return
	}

	if value == "" {
		b.WriteInt(0)
		return
	}

	strBytes := []byte(value)
	length := int32(len(strBytes))

	if length > maxStringLength {
		slog.Error("will not write string because it is too long", "size", length)
		b.WriteInt(-1)

		return
	}

	b.WriteInt(length)

	if length > 0 {
		b.writeBytesInternal(strBytes)
	}
}

func (b *ByteStream) WriteVInt(data VInt) {
	b.bitOffset = 0

	var final []byte
	d := int32(data)

	if d < 0 {
		if d >= -63 {
			final = append(final, byte((d&0x3F)|0x40))
		} else if d >= -8191 {
			final = append(final, byte((d&0x3F)|0xC0))
			final = append(final, byte((uint32(d>>6))&0x7F))
		} else if d >= -1048575 {
			final = append(final, byte((d&0x3F)|0xC0))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>13))&0x7F))
		} else if d >= -134217727 {
			final = append(final, byte((d&0x3F)|0xC0))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>13))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>20))&0x7F))
		} else {
			final = append(final, byte((d&0x3F)|0xC0))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>13))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>20))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>27))&0x0F))
		}
	} else {
		if d <= 63 {
			final = append(final, byte(d&0x3F))
		} else if d <= 8191 {
			final = append(final, byte((d&0x3F)|0x80))
			final = append(final, byte((uint32(d>>6))&0x7F))
		} else if d <= 1048575 {
			final = append(final, byte((d&0x3F)|0x80))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>13))&0x7F))
		} else if d <= 134217727 {
			final = append(final, byte((d&0x3F)|0x80))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>13))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>20))&0x7F))
		} else {
			final = append(final, byte((d&0x3F)|0x80))
			final = append(final, byte(((uint32(d>>6))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>13))&0x7F)|0x80))
			final = append(final, byte(((uint32(d>>20))&0x7F)|0x80))
			final = append(final, byte((uint32(d>>27))&0x0F))
		}
	}

	b.writeBytesInternal(final)
}

func (b *ByteStream) WriteDataRef(value DataRef) {
	b.WriteVInt(VInt(value.F))

	if value.F != 0 {
		b.WriteVInt(VInt(value.S))
	}
}

func (b *ByteStream) WriteScId(value ScId) {
	b.WriteDataRef(DataRef(value))
}

func (b *ByteStream) WriteLogicLong(value LogicLong) {
	b.bitOffset = 0

	b.WriteVInt(VInt(value.F))
	b.WriteVInt(VInt(value.S))
}

func (b *ByteStream) WriteArrayVInt(value []VInt) {
	b.bitOffset = 0

	b.WriteVInt(VInt(len(value)))

	for _, element := range value {
		b.WriteVInt(element)
	}
}

func (b *ByteStream) WriteByte(value byte) {
	b.buffer = append(b.buffer, value&0xFF)
}

// --- Utility --- //

func (b *ByteStream) Write(value interface{}) {
	switch v := value.(type) {
	case int32:
		b.WriteInt(v)
	case int:
		b.WriteInt(int32(v))
	case VInt:
		b.WriteVInt(v)
	case bool:
		b.WriteBool(v)
	case string:
		b.WriteString(v)
	case DataRef:
		b.WriteDataRef(v)
	case ScId:
		b.WriteScId(v)
	case LogicLong:
		b.WriteLogicLong(v)
	case []VInt:
		b.WriteArrayVInt(v)
	case byte:
		b.WriteByte(v)
	default:
		slog.Warn("faled to write(x): unsupported type!", "value", v)
	}
}

func (b *ByteStream) Buffer() []byte {
	return b.buffer
}

func (b *ByteStream) Offset() int {
	return b.offset
}
