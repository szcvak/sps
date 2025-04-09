package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

const (
	defaultInitialCapacity = 128
	maxStringLength        = 900_000

	EmptyString = "<empty>"
)

var (
	ErrBufferTooSmall       = errors.New("buffer too small for read operation")
	ErrStringTooLong        = errors.New("string is too long")
	ErrStringNegativeLength = errors.New("string has negative length")
)

type VInt int32

type ByteStream struct {
	buffer []byte

	offset    int
	bitOffset int
}

func NewByteStream(initialBuffer []byte) *ByteStream {
	var capacity int = defaultInitialCapacity

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

func (b *ByteStream) Destroy() {
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

func (b *ByteStream) ensureWriteCapacity(n int) {
	needed := b.offset + n

	if needed > cap(b.buffer) {
		newCap := cap(b.buffer) * 2

		if newCap < needed {
			newCap = needed
		}

		newBuffer := make([]byte, len(b.buffer), newCap)
		copy(newBuffer, b.buffer)

		b.buffer = newBuffer
	}
}

// --- Read operations --- //

func (b *ByteStream) ReadBool() (bool, error) {
	if b.bitOffset == 0 {
		if err := b.ensureReadCapacity(1); err != nil {
			return false, fmt.Errorf("failed to read boolean: %w", err)
		}
	}

	byteIndex := b.offset
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
		return "", fmt.Errorf("failed to read string: %w", err)
	}

	if length == -1 || length == 0 {
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
		return "", fmt.Errorf("failed to read string: %w", err)
	}

	strBytes := b.buffer[b.offset : b.offset+int(length)]

	if !utf8.Valid(strBytes) {
		b.offset += int(length)
		return "", fmt.Errorf("failed to read string: invalid utf8")
	}

	b.offset += int(length)
	result := string(strBytes)

	return result, nil
}

func (b *ByteStream) ReadVInt() (int32, error) {
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
			signBitPos := uint(bytesRead * 7)

			if signBitPos < 32 && (result>>(signBitPos-1))&1 != 0 {
				mask := int64(-1) << signBitPos
				result |= mask
			}

			if result < math.MinInt32 || result > math.MaxInt32 {
				return 0, fmt.Errorf("failed to read variable-length int: value %d underflows int32", result)
			}

			return int32(result), nil
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
		b.offset++
	}

	last := len(b.buffer) - 1

	if value {
		b.buffer[last] |= 1 << (b.bitOffset & 31)
	}

	b.bitOffset = b.bitOffset + 1&7
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

	strBytes := []byte(value)
	length := int32(len(strBytes))

	if length > maxStringLength {
		b.WriteInt(-1)
		return
	}

	b.WriteInt(length)

	if length > 0 {
		b.writeBytesInternal(strBytes)
	}
}

func (b *ByteStream) WriteVInt(value VInt) {
	b.bitOffset = 0

	uval := uint32(value)

	var buf [5]byte
	idx := 0

	zigzag := (uval << 1) ^ uint32(int32(uval)>>31)
	result := zigzag

	for {
		temp := byte(result & 0x7F)
		result >>= 7

		if result != 0 {
			temp |= 0x80
		}

		buf[idx] = temp
		idx++

		if result == 0 {
			break
		}
	}

	b.writeBytesInternal(buf[:idx])
}

// --- Utility --- //

func (b *ByteStream) Write(value interface{}) {
	switch v := any(value).(type) {
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
	default:
		fmt.Printf("can't write bytes because the type is not supported (%w)\n", v)
	}
}

func (b *ByteStream) Buffer() []byte {
	return b.buffer
}
