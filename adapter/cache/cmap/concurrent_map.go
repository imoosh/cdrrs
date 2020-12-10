package concurrent_map

import (
    "github.com/RussellLuo/timingwheel"
    "sync"
    "time"
)

// ConcurrentMap is a thread safe map collection with better performance.
// The backend map entries are separated into the different partitions.
// Threads can access the different partitions safely without lock.
type ConcurrentMap struct {
    partitions    []*innerMap
    numOfBlockets int
}

// Partitionable is the interface which should be implemented by key type.
// It is to define how to partition the entries.
type Partitionable interface {
    // Value is raw value of the key
    Value() interface{}

    // PartitionKey is used for getting the partition to store the entry with the key.
    // E.g. the key's hash could be used as its PartitionKey
    // The partition for the key is partitions[(PartitionKey % m.numOfBlockets)]
    //
    // 1 Why not provide the default hash function for partition?
    // Ans: As you known, the partition solution would impact the performance significantly.
    // The proper partition solution balances the access to the different partitions and
    // avoid of the hot partition. The access mode highly relates to your business.
    // So, the better partition solution would just be designed according to your business.
    PartitionKey() int64
}

type MapValue struct {
    Raw   interface{}
    Timer *timingwheel.Timer
}

type innerMap struct {
    m map[interface{}]interface{}

    // 每个map的哈希桶下创建一个时间轮，降低锁粒度
    tw   *timingwheel.TimingWheel
    lock sync.RWMutex

    expire time.Duration
}

func createInnerMap(expire time.Duration) *innerMap {
    // 刻度为1秒，轮槽为1800个
    tw := timingwheel.NewTimingWheel(time.Second, 1800)
    tw.Start()

    return &innerMap{
        m:      make(map[interface{}]interface{}),
        tw:     tw,
        expire: expire,
    }
}

func (im *innerMap) get(key Partitionable) (interface{}, bool) {
    keyVal := key.Value()
    im.lock.RLock()
    v, ok := im.m[keyVal]
    im.lock.RUnlock()
    if !ok {
        return nil, ok
    }
    return v.(MapValue).Raw, ok
}

// add by wayne
func (im *innerMap) setex(key Partitionable, v interface{}) (interface{}, bool) {
    var ok bool
    var val interface{}
    keyVal := key.Value()

    im.lock.Lock()
    if val, ok = im.m[keyVal]; !ok {
        // 没有缓存，直接存入
        im.m[keyVal] = MapValue{v, im.tw.AfterFunc(im.expire, func() { im.del(key) })}
        im.lock.Unlock()

        return nil, true
    } else {
        // 已存在，缓存失败, invite遇到bye，或bye遇到invite，立即在锁内删除
        timer := val.(MapValue).Timer
        if val != v {
            delete(im.m, keyVal)
        }
        im.lock.Unlock()
        if timer != nil {
            timer.Stop()
        }
        return val.(MapValue).Raw, false
    }
}

func (im *innerMap) set(key Partitionable, v interface{}) {
    keyVal := key.Value()
    im.lock.Lock()
    im.m[keyVal] = MapValue{v, nil}
    im.lock.Unlock()
}

func (im *innerMap) del(key Partitionable) {
    keyVal := key.Value()
    im.lock.Lock()
    delete(im.m, keyVal)
    im.lock.Unlock()
}

// CreateConcurrentMap is to create a ConcurrentMap with the setting number of the partitions
func CreateConcurrentMap(numOfPartitions int, expire time.Duration) *ConcurrentMap {
    var partitions []*innerMap
    for i := 0; i < numOfPartitions; i++ {
        partitions = append(partitions, createInnerMap(expire))
    }
    return &ConcurrentMap{partitions, numOfPartitions}
}

func (m *ConcurrentMap) getPartition(key Partitionable) *innerMap {
    partitionID := key.PartitionKey() % int64(m.numOfBlockets)
    return (*innerMap)(m.partitions[partitionID])
}

// Get is to get the value by the key
func (m *ConcurrentMap) Get(key Partitionable) (interface{}, bool) {
    return m.getPartition(key).get(key)
}

// SetEx is to store the KV entry to the map
func (m *ConcurrentMap) Set(key Partitionable, v interface{}) {
    im := m.getPartition(key)
    im.set(key, v)
}

// Del is to delete the entries by the key
func (m *ConcurrentMap) Del(key Partitionable) {
    im := m.getPartition(key)
    im.del(key)
}

func (m *ConcurrentMap) SetEx(key Partitionable, v interface{}) (interface{}, bool) {
    im := m.getPartition(key)
    return im.setex(key, v)
}
