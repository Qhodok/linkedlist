package List

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
)

type EVENT int

const (
	ADD EVENT = iota
	UPDATE
	DELETE
)

type Event[T any] struct {
	Key     string
	Content T
	Event   EVENT
}

type Storage interface {
	//SetAddAndDeleteFunction(add AddElement, del DeleteElement)
	Add(key string, data []byte)
	Delete(key string)
	Update(key string, data []byte)
	Get(key string) interface{}
}

type Component struct {
	Key  string     `json:"key"`
	Next *Component `json:"next"`
	Prev *Component `json:"prev"`
}

type GenericChan[T any] chan Event[T]

type LinkedList[Type any] struct {
	container    sync.Map
	queue        map[string]*Component
	head         *Component
	tail         *Component
	locker       sync.Mutex
	eventChannel chan Event[Type]
	storage      Storage
	totalAccess  int
}

func NewPointerList[Type any]() (linkedList *LinkedList[Type]) {
	linkedList = &LinkedList[Type]{
		queue: make(map[string]*Component),
		head:  nil,
		tail:  nil,
	}
	return
}

func NewList[Type any]() (linkedList LinkedList[Type]) {
	linkedList = LinkedList[Type]{
		queue: make(map[string]*Component),
		head:  nil,
		tail:  nil,
	}
	return
}

func NewListWithStorage[Type any](storage Storage) (linkedList LinkedList[Type]) {
	linkedList = LinkedList[Type]{
		queue:   make(map[string]*Component),
		head:    nil,
		tail:    nil,
		storage: storage,
	}
	return
}
func NewPointerListWithStorage[Type any](storage Storage) (linkedList *LinkedList[Type]) {
	linkedList = &LinkedList[Type]{
		queue:   make(map[string]*Component),
		head:    nil,
		tail:    nil,
		storage: storage,
	}
	return
}

func (this *LinkedList[Type]) CreateEventListener(buffer int) (int, chan Event[Type]) {
	if this.eventChannel == nil {
		this.eventChannel = make(chan Event[Type], buffer)
	}
	return cap(this.eventChannel), this.eventChannel
}

func (this *LinkedList[Type]) AddLast(key string, data Type) (error error) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if _, ok := this.container.Load(key); ok {
		error = errors.New("duplicate key")
	} else {
		temp := &Component{Key: key}
		this.queue[key] = temp
		this.container.Store(key, data)
		if this.head == nil {
			this.head = temp
			this.tail = temp
		} else {
			this.tail.Next = temp
			temp.Prev = this.tail
			this.tail = temp
		}
		this.handleEvent(ADD, key)
	}
	return
}

func (this *LinkedList[T]) AddLastOnExistIgnore(key string, data T) (result bool) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if _, ok := this.container.Load(key); ok {
		result = false
	} else {
		temp := &Component{Key: key}
		this.queue[key] = temp
		this.container.Store(key, data)
		if this.head == nil {
			this.head = temp
			this.tail = temp
		} else {
			this.tail.Next = temp
			temp.Prev = this.tail
			this.tail = temp
		}
		result = true
		this.handleEvent(ADD, key)
	}
	return
}

func (this *LinkedList[T]) Update(key string, data T) error {
	this.locker.Lock()
	defer this.locker.Unlock()
	if oldData, ok := this.container.Load(key); ok {
		if !reflect.DeepEqual(oldData, data) {
			reflectNewData := reflect.ValueOf(data)
			if reflectNewData.Kind() == reflect.Ptr {
				starX := reflectNewData.Elem()
				y := reflect.New(starX.Type())
				starY := y.Elem()
				starY.Set(starX)
				reflect.ValueOf(oldData).Elem().Set(y.Elem())
			} else {
				this.container.Store(key, data)
			}
			this.handleEvent(UPDATE, key)
		}
		return nil
	} else {
		return errors.New("data not found")
	}
}

func (this *LinkedList[T]) AddLastOrUpdate(key string, data T, deepEqual ...bool) (error error, event EVENT) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if oldData, ok := this.container.Load(key); ok {
		if len(deepEqual) > 0 && deepEqual[0] {
			if !reflect.DeepEqual(oldData, data) {
				reflectNewData := reflect.ValueOf(data)
				if reflectNewData.Kind() == reflect.Ptr {
					starX := reflectNewData.Elem()
					y := reflect.New(starX.Type())
					starY := y.Elem()
					starY.Set(starX)
					reflect.ValueOf(oldData).Elem().Set(y.Elem())
				} else {
					this.container.Store(key, data)
				}
				this.handleEvent(UPDATE, key)
				event = UPDATE
			}
		} else {
			this.container.Store(key, data)
			this.handleEvent(UPDATE, key)
			event = UPDATE
		}
	} else {
		temp := &Component{Key: key}
		this.queue[key] = temp
		this.container.Store(key, data)
		if this.head == nil {
			this.head = temp
			this.tail = temp
		} else {
			this.tail.Next = temp
			temp.Prev = this.tail
			this.tail = temp
		}
		this.handleEvent(ADD, key)
		event = ADD
	}
	return
}

func (this *LinkedList[T]) AddLastOrUpdateSkipStorage(key string, data T, deepEqual ...bool) (error error, event EVENT) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if oldData, ok := this.container.Load(key); ok {
		if len(deepEqual) > 0 && deepEqual[0] {
			if !reflect.DeepEqual(oldData, data) {
				reflectNewData := reflect.ValueOf(data)
				if reflectNewData.Kind() == reflect.Ptr {
					starX := reflectNewData.Elem()
					y := reflect.New(starX.Type())
					starY := y.Elem()
					starY.Set(starX)
					reflect.ValueOf(oldData).Elem().Set(y.Elem())
				} else {
					this.container.Store(key, data)
				}
				this.handleEvent(UPDATE, key, true)
				event = UPDATE
			}
		} else {
			this.container.Store(key, data)
			this.handleEvent(UPDATE, key, true)
			event = UPDATE
		}
	} else {
		temp := &Component{Key: key}
		this.queue[key] = temp
		this.container.Store(key, data)
		if this.head == nil {
			this.head = temp
			this.tail = temp
		} else {
			this.tail.Next = temp
			temp.Prev = this.tail
			this.tail = temp
		}
		this.handleEvent(ADD, key, true)
		event = ADD
	}
	return
}

func (this *LinkedList[Type]) AddFirst(key string, data Type) (error error) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if _, ok := this.container.Load(key); ok {
		error = errors.New("duplicate key")
	} else {
		temp := &Component{Key: key}
		this.queue[key] = temp
		this.container.Store(key, data)
		if this.head == nil {
			this.head = temp
			this.tail = temp
		} else {
			this.head.Prev = temp
			temp.Next = this.head
			this.head = temp
		}
		this.handleEvent(ADD, key)
	}
	return
}

func (this *LinkedList[Type]) AddAfter(key string, target string, data Type) (found bool, error error) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if _, found = this.container.Load(key); found {
		error = errors.New("duplicate key")
	} else {
		var container *Component
		if container, found = this.queue[target]; found {
			temp := &Component{Key: key, Prev: container}
			this.queue[key] = temp
			this.container.Store(key, data)
			if container.Next == nil {
				container.Next = temp
				this.tail = temp
			} else {
				container.Next.Prev = temp
				temp.Next = container.Next
				container.Next = temp
			}
			this.handleEvent(ADD, key)
		} else {
			error = errors.New("target not found")
		}
	}
	return
}

func (this *LinkedList[Type]) AddBefore(key string, target string, data Type) (found bool, error error) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if _, found = this.container.Load(key); found {
		error = errors.New("duplicate key")
	} else {
		var container *Component
		if container, found = this.queue[target]; found {
			temp := &Component{Key: key, Next: container}
			this.queue[key] = temp
			this.container.Store(key, data)
			if container.Prev == nil {
				container.Prev = temp
				this.head = temp
			} else {
				temp.Prev = container.Prev
				container.Prev.Next = temp
				container.Prev = temp
			}
			this.handleEvent(ADD, key)
		} else {
			error = errors.New("target not found")
		}
	}
	return
}

// delete
func (this *LinkedList[Type]) RemoveFirst() (key string, value Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.head != nil {
		data := this.head
		this.head = this.head.Next
		if this.head != nil {
			this.head.Prev = nil
		}
		key = data.Key
		load, ok := this.container.Load(key)
		if ok {
			value = load.(Type)
		}
		this.container.Delete(key)
		delete(this.queue, data.Key)
		this.handleEvent(DELETE, key)
	}
	return
}

func (this *LinkedList[Type]) RemoveLast() (key string, value Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.tail != nil {
		data := this.tail
		this.tail = this.tail.Prev
		if this.tail != nil {
			this.tail.Next = nil
		}
		key = data.Key
		load, ok := this.container.Load(key)
		if ok {
			value = load.(Type)
		}
		this.container.Delete(key)
		delete(this.queue, data.Key)
		this.handleEvent(DELETE, key)
	}
	return
}

func (this *LinkedList[Type]) RemoveAfter(targetList string) (key string, value Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if target, ok := this.queue[targetList]; ok {
		if target.Next != nil {
			data := target.Next
			target.Next = data.Next
			if data.Next != nil {
				target.Next.Prev = target
			} else {
				this.tail = target
			}
			key = data.Key
			load, ok := this.container.Load(key)
			if ok {
				value = load.(Type)
			}
			this.container.Delete(key)
			delete(this.queue, data.Key)
			this.handleEvent(DELETE, key)
		}
	}
	return
}

func (this *LinkedList[Type]) RemoveBefore(targetList string) (key string, value Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if target, ok := this.queue[targetList]; ok {
		if target.Prev != nil {
			data := target.Prev
			target.Prev = data.Prev
			if data.Prev != nil {
				target.Prev.Next = target
			} else {
				this.head = target
			}
			key = data.Key
			load, ok := this.container.Load(key)
			if ok {
				value = load.(Type)
			}
			this.container.Delete(key)
			delete(this.queue, data.Key)
			this.handleEvent(DELETE, key)
		}
	}
	return
}

func (this *LinkedList[Type]) Remove(target string) (exist bool, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if data, ok := this.queue[target]; ok {
		//only one data
		if data.Next == nil && data.Prev == nil {
			this.head = nil
			this.tail = nil
			//data is head
		} else if data.Next != nil && data.Prev == nil {
			this.head = data.Next
			this.head.Prev = nil
			//data is tail
		} else if data.Prev != nil && data.Next == nil {
			this.tail = data.Prev
			this.tail.Next = nil
			// data in middle
		} else {
			data.Prev.Next = data.Next
			data.Next.Prev = data.Prev
		}
		exist = true
		load, ok := this.container.Load(data.Key)
		if ok {
			element = load.(Type)
		}
		this.container.Delete(data.Key)
		delete(this.queue, data.Key)
		go this.handleEvent(DELETE, data.Key)
	}
	return
}

func (this *LinkedList[Type]) Find(target string) (result bool, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if data, ok := this.container.Load(target); ok {
		element = data.(Type)
		result = true
	}
	return
}

func (this *LinkedList[Type]) Next(target string) (key string, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if data, ok := this.queue[target]; ok {
		if data.Next != nil {
			data = data.Next
			key = data.Key
			load, ok := this.container.Load(key)
			if ok {
				element = load.(Type)
			}
			return
		}
	}
	return
}

func (this *LinkedList[Type]) Prev(target string) (key string, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if data, ok := this.queue[target]; ok {
		if data.Prev != nil {
			data = data.Next
			key = data.Key
			load, ok := this.container.Load(key)
			if ok {
				element = load.(Type)
			}
		}
	}
	return
}

func (this *LinkedList[Type]) Head() (key string, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.head != nil {
		key = this.head.Key
		load, ok := this.container.Load(key)
		if ok {
			element = load.(Type)
		}
	}
	return
}

func (this *LinkedList[Type]) Tail() (key string, element Type) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.tail != nil {
		key = this.tail.Key
		load, ok := this.container.Load(key)
		if ok {
			element = load.(Type)
		}
	}
	return
}

func (this *LinkedList[Type]) Size() int {
	this.locker.Lock()
	defer this.locker.Unlock()
	return len(this.queue)
}

func (this *LinkedList[Type]) handleEvent(event EVENT, key string, skipStorage ...bool) {
	var data Type
	temp, ok := this.container.Load(key)
	if ok {
		data = temp.(Type)
	}
	if this.eventChannel != nil {
		lenChan := len(this.eventChannel)
		capChan := cap(this.eventChannel)
		e := Event[Type]{
			Key:     key,
			Content: data,
			Event:   event,
		}
		if lenChan > (capChan*9)/10 {
			go func() {
				this.eventChannel <- e
			}()
		} else {
			this.eventChannel <- e
		}
	}

	if this.storage != nil && len(skipStorage) == 0 {
		switch event {
		case ADD:
			marshal, _ := json.Marshal(temp)
			this.storage.Add(key, marshal)
		case UPDATE:
			marshal, _ := json.Marshal(temp)
			this.storage.Update(key, marshal)
		case DELETE:
			this.storage.Delete(key)
		}
	}
}

func (this *LinkedList[Type]) Contents() (result map[string]Type) {
	result = make(map[string]Type)
	this.container.Range(func(key, value any) bool {
		result[key.(string)] = value.(Type)
		return true
	})
	return
}
