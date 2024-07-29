package bitmap

type Bitmap struct {
	bits []byte
	size int
}

// todo: 防止碰撞 (用线性探测或拉链法)

func NewBitmap(size int) *Bitmap {
	if size == 0 {
		size = 128
	}
	return &Bitmap{
		bits: make([]byte, size),
		size: size * 8,
	}
}

func (b *Bitmap) Set(id string) {
	// 获取索引下标
	idx := hash(id) % b.size
	// 计算在哪个byte
	byteIdx := idx / 8
	// 计算在这个byte中的哪个bit位置
	bitIdx := idx % 8
	b.bits[byteIdx] |= 1 << bitIdx
}

// 验证
func (b *Bitmap) IsSet(id string) bool {
	idx := hash(id) % b.size
	// 计算在那个byte
	byteIdx := idx / 8
	// 在这个byte中的那个bit位置
	bitIdx := idx % 8
	return (b.bits[byteIdx] & (1 << bitIdx)) != 0
}

// 导出
func (b *Bitmap) Export() []byte {
	return b.bits
}

// 加载
func Load(bits []byte) *Bitmap {
	if len(bits) == 0 {
		return NewBitmap(0)
	}

	return &Bitmap{
		bits: bits,
		size: len(bits) * 8,
	}
}

func hash(id string) int {
	seed := 131313
	hash := 0
	for _, c := range id {
		hash = hash*seed + int(c)
	}
	return hash & 0x7FFFFFFF
}
