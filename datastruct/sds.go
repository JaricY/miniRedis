package datastruct

import (
	"bytes"
	"unsafe"
)

const SDS_MAX_PREALLOC = 1024 * 1024

type SDS struct {
	buf  []byte
	len  int32
	free int32
}

func NewSDS(str string) *SDS { // 创建新的sds对象
	s := &SDS{}
	s.NewSDS(str)
	return s
}

func (s *SDS) NewSDS(str string) {
	length := int32(len(str))
	if length <= 1024 {
		//默认创建1KB大小的缓冲
		s.buf = make([]byte, 1024)
		s.buf = append(s.buf, str...)
		s.len = length
		s.free = 1024 - length
	} else {
		//如果大于1KB，则不预留空间
		s.buf = []byte(str)
		s.len = length
		s.free = 0
	}
}

func (s *SDS) Len() int32 { //返回已使用的空间字节数
	return s.len
}

func (s *SDS) Free() int32 { //返回未使用的空间字节数
	return s.free
}

func (s *SDS) Grow(len int32) { //扩容，参数len表示当前所需扩增的长度
	free := s.Free()

	if free > len {
		return
	}
	var len1 int32
	//len1是扩增后的长度
	if len < SDS_MAX_PREALLOC {
		//如果所需扩增的长度小于1M
		len1 = s.Len() + len

		if len1 < SDS_MAX_PREALLOC {
			//如果扩增后的长度小于1M
			len1 = len1 * 2
		} else {
			len1 += SDS_MAX_PREALLOC
		}
	} else {
		len1 = SDS_MAX_PREALLOC + len
	}

	newbuf := make([]byte, len1)

	copy(newbuf, s.buf)

	s.buf = newbuf

	// 先做扩增，具体的长度还没确定
	s.free = len1 - s.Len()
}

func (s *SDS) Append(str string) {
	bytes := StringToBytes(str)
	length := int32(len(bytes))
	free := s.Free()
	if length > free {
		s.Grow(length)
	}
	copy(s.buf, bytes)
	s.len += length

}

func (s *SDS) String() string {
	if s == nil {
		return ""
	}
	return BytesToString(s.buf)
}

// TODO
func BytesToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func StringToBytes(str string) []byte {
	sh := (*[2]uintptr)(unsafe.Pointer(&str))
	return BytesFromStringHeader((*[3]uintptr)(unsafe.Pointer(sh)))
}

func BytesFromStringHeader(hdr *[3]uintptr) []byte {
	b := make([]byte, hdr[1])
	copy(b, *(*[]byte)(unsafe.Pointer(hdr[0])))
	return b
}

func SDSCmp(s1 SDS, s2 SDS) int32 {
	len1 := s1.Len()
	len2 := s2.Len()
	minlen := len1

	if len1 > len2 {
		minlen = len2
	}

	cmp := bytes.Compare(s1.buf[:minlen], s2.buf[:minlen])

	if cmp == 0 {
		return len1 - len2
	}

	return int32(cmp)
}
