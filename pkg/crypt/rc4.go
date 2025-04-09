package crypt

type Rc4 struct {
	s []byte
	i byte
	j byte
}

func NewRc4(key []byte) *Rc4 {
	s := make([]byte, 256)

	for i := 0; i < 256; i++ {
		s[i] = uint8(i)
	}

	var j byte = 0

	for i := 0; i < 256; i++ {
		j = j + s[i] + key[i%len(key)]
		s[i], s[j] = s[j], s[i]
	}

	return &Rc4{
		s: s,
		j: 0,
		i: 0,
	}
}

func (r *Rc4) Process(data []byte) {
	for k := range data {
		r.i++
		r.j += r.s[r.i]

		r.s[r.i], r.s[r.j] = r.s[r.j], r.s[r.i]
		keystream := r.s[r.s[r.i]+r.s[r.j]]

		data[k] ^= keystream
	}
}
