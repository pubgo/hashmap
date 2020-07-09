package hashmap

type Hashmap interface {
	Size() int
	Count() int
	Del(key []byte) (err error)
	Get(key []byte) (val []byte, err error)
	Set(key []byte, val []byte) (err error)
}
