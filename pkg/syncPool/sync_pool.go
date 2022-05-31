package syncPool

import "sync"

type SyncPool struct {
	pool     sync.Pool
	newFunc  func() interface{}
	initFunc func(interface{})
}

// NewSyncPool 创建对象池
// newFunc：创建对象的方法
// init： 对象被创建后，返回之前，调用该方法初始化对象
func NewSyncPool(newFunc func() interface{}, init func(interface{})) *SyncPool {
	return &SyncPool{
		newFunc:  newFunc,
		initFunc: init,
		pool: sync.Pool{
			New: newFunc,
		},
	}
}

// Get 获取对象
func (that *SyncPool) Get() interface{} {
	var object = that.pool.Get()
	if that.initFunc != nil {
		that.initFunc(object)
	}
	return object
}

// Put 把对象放回对象池
func (that *SyncPool) Put(value interface{}) {
	that.pool.Put(value)
}
