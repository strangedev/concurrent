package concurrent

import (
	"sync"
)

type entry struct {
	Key 	int
	Data	interface{}
	Done 	*chan bool
}

type mapWriteChans struct {
	Stop	chan bool
	Upsert	chan entry
	Remove	chan entry
}

type MapWriter interface {
	Stop()
	Upsert(k int, data interface{}) (done chan bool)
	Remove(k int) (done chan bool)
}

func (c *ConcurrentHashMap) Stop() {
	c.writer.Stop<- true
}

func (c* ConcurrentHashMap) Upsert(k int, data interface{}) (done chan bool) {
	done = make(chan bool, 1)
	c.writer.Upsert<- entry{
		Key:		k,
		Data:		data,
		Done:		&done,
	}
	return 
}

func (c* ConcurrentHashMap) Remove(k int) (done chan bool) {
	done = make(chan bool, 1)
	c.writer.Remove<- entry{
		Key:		k,
		Data:		nil,
		Done:		&done,
	}
	return
}

type MapReader interface {
	Get(k int) (e entry, ok bool)
}

type ConcurrentHashMap struct {
	writer 		*mapWriteChans
	writeMap	map[int]interface{}
	readMap		map[int]interface{}
	readLock	sync.RWMutex
}

func (c *ConcurrentHashMap) Get(k int) (data interface{}, ok bool) {
	c.readLock.RLock()
	data, ok = c.readMap[k]
	c.readLock.RUnlock()
	return
}

func NewHashMap() *ConcurrentHashMap {
	writer := &mapWriteChans{
		Stop: 	make(chan bool),
		Upsert: make(chan entry, 8),
		Remove: make(chan entry, 8),
	}
	m := &ConcurrentHashMap{
		writer:		writer,
		readMap:	make(map[int]interface{}),
		writeMap:	make(map[int]interface{}),
	}
	go m.run()
	return m
}

func (c *ConcurrentHashMap)run() {
	for {
		select {
		case _ = <- c.writer.Stop:
			close(c.writer.Stop)
			close(c.writer.Upsert)
			close(c.writer.Remove)
			return

		case entry := <- c.writer.Upsert:
			c.writeMap[entry.Key] = entry.Data
			c.readLock.Lock()
			c.readMap, c.writeMap = c.writeMap, c.readMap
			c.readLock.Unlock()
			c.writeMap[entry.Key] = entry.Data
			*(entry.Done)<- true

		case entry := <- c.writer.Remove:
			_, isInMap := c.readMap[entry.Key]
			delete(c.writeMap, entry.Key)
			if isInMap {
				c.readLock.Lock()
				c.readMap, c.writeMap = c.writeMap, c.readMap
				c.readLock.Unlock()
				delete(c.writeMap, entry.Key)
			}
			*(entry.Done)<- true
		}
	}
}