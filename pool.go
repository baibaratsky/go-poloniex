package poloniex

type keyPool struct {
	keys chan *Key
}

func (keyPool keyPool) Put(key *Key) {
	keyPool.keys <- key
}

func (keyPool keyPool) Get() *Key {
	return <-keyPool.keys
}
