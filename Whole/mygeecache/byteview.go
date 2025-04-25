package mygeecache

// A ByteView holds an immutable（持久化） view of bytes.

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int { return len(v.b) }

func (v ByteView) ByteSlice() []byte { return cloneBytes(v) }

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(v ByteView) []byte {
	c := make([]byte, len(v.b))
	copy(c, v.b)
	return c
}
