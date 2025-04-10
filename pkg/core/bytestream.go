package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
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

	shift := uint(0)
	result := int64(0)
	bytesRead := 0

	for bytesRead < 5 {
		if b.offset >= len(b.buffer) {
			return 0, fmt.Errorf("failed to read variable-length int: %w", io.EOF)
		}

		bVal := b.buffer[b.offset]
		b.offset++
		bytesRead++

		data := int64(bVal & 0x7F)
		result |= data << shift
		shift += 7

		if (bVal & 0x80) == 0 {
			signBitMask := int64(1) << (shift - 1)

			if (result & signBitMask) != 0 {
				mask := int64(-1) << shift
				result |= mask
			}

			if result < math.MinInt32 || result > math.MaxInt32 {
				fmt.Printf("result is outside the bounds of int32: %d", result)
			}

			return VInt(result), nil
		}
	}

	return 0, fmt.Errorf("failed to read variable-length int: read 5 bytes without termination")
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
		_, _ = fmt.Fprintf(os.Stderr, "will not write string because it is too long\n")
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
	b.buffer = append(b.buffer, value & 0xFF)
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
		_, _ = fmt.Fprintf(os.Stderr, "write(x) unsupported type (%T)\n", v)
	}
}

func (b *ByteStream) Buffer() []byte {
	return b.buffer
}

func (b *ByteStream) Offset() int {
	return b.offset
}