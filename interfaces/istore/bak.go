package istore

// TODO:自动主备切换策略

func NewBakStore(main Store, baks []Store) Store {
	return &BakStore{
		baks: baks,
		main: main,
	}
}

type BakStore struct {
	baks []Store
	main Store
}

func (b *BakStore) Delete(key string) (err error) {
	err = b.main.Delete(key)
	b.take(err)
	return err
}

func (b *BakStore) Get(key string) (value string, err error) {
	value, err = b.main.Get(key)
	b.take(err)
	return
}

func (b *BakStore) Set(key, value string) (err error) {
	err = b.main.Set(key, value)
	b.take(err)
	return
}

func (b *BakStore) Add(key, value string) (err error) {
	err = b.main.Add(key, value)
	b.take(err)
	return
}

func (b *BakStore) take(err error) {
	if err != nil {
		return
	}
}

var _ Store = &BakStore{}
