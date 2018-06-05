package heap

import (
	"reflect"
	"errors"
	coheap "container/heap"
)

type Indexer interface {
	GetIndex() int
	SetIndex(int)
}

type IndexMixin struct {
	index int
}

func (i IndexMixin) GetIndex() int {
	return i.index
}

func (i *IndexMixin) SetIndex(index int) {
	i.index = index
}

type Heap struct {
	objects []reflect.Value

	cmpFn reflect.Value
	dataType reflect.Type

	indexer bool

	lookup map[reflect.Value]int
}

func NewHeap(compareFn interface{}) (*Heap, error) {
	h := &Heap{
		objects:make([]reflect.Value, 0),
		lookup:make(map[reflect.Value]int),
	}
	if err := h.checkAndSetFn(compareFn); nil != err {
		return nil, err
	}

	return h, nil
}

func MustHeap(compareFn interface{}) *Heap {
	h, err := NewHeap(compareFn)
	if nil != err {
		panic(err)
	}
	coheap.Init(h)
	return h
}

func (h *Heap) checkAndSetFn(compareFn interface{}) error {
	to := reflect.TypeOf(compareFn)
	if to.Kind() != reflect.Func {
		return errors.New("not a function")
	}
	if to.NumOut() != 1 {
		return errors.New("invalid amount of return params")
	}

	if to.Out(0).Kind() != reflect.Bool {
		return errors.New("return value must be bool")
	}

	if to.NumIn() != 2 {
		return errors.New("expected exactly two parameters, only one gotten")
	}

	if to.In(0) != to.In(1) {
		return errors.New("both input parameters of the function must be of the same type")
	}

	h.cmpFn = reflect.ValueOf(compareFn)

	h.dataType = to.In(0)
	if h.dataType.Kind() != reflect.Ptr {
		return errors.New("elems must have a pointer receiver")
	}
	if h.dataType.Implements(reflect.TypeOf((*Indexer)(nil)).Elem()) {
		h.indexer = true
	}

	return nil
}

func (h Heap) Less(i, j int) bool {
	return h.cmpFn.Call([]reflect.Value{h.objects[i], h.objects[j]})[0].Interface().(bool)
}

func (h Heap) Swap(i, j int) {
	if h.indexer {
		h.objects[i].Interface().(Indexer).SetIndex(j)
		h.objects[j].Interface().(Indexer).SetIndex(i)
	}
	h.lookup[h.objects[i]] = j
	h.lookup[h.objects[j]] = i
	h.objects[i], h.objects[j] = h.objects[j], h.objects[i]
}

func (h Heap) Len() int {
	return len(h.objects)
}

func (h *Heap) Push(i interface{}) {
	if reflect.TypeOf(i) != h.dataType {
		panic("tried to put invalid type")
	}
	val := reflect.ValueOf(i)
	h.lookup[val] = len(h.objects)
	h.objects = append(h.objects, val)
}

func (h *Heap) Pop() interface{} {
	length := len(h.objects)
	ret := h.objects[length - 1].Interface()
	h.objects = h.objects[:length-1]
	return ret
}

func (h *Heap) Put(i interface{}) {
	coheap.Push(h, i)
}

func (h *Heap) Get(i interface{}) {
	if reflect.TypeOf(i) != h.dataType {
		panic("bad target type")
	}
	ret := coheap.Pop(h)
	reflect.ValueOf(i).Elem().Set(reflect.ValueOf(ret).Elem())
}

func (h *Heap) Peek(i interface{}) {
	if reflect.TypeOf(i) != h.dataType {
		panic("bad target type")
	}
	ret := h.objects[0]
	reflect.ValueOf(i).Elem().Set(reflect.ValueOf(ret).Elem())
}

func (h *Heap) DeleteElem(i interface{}) bool {
	v := reflect.ValueOf(i)
	index, ok := h.lookup[v]
	if !ok {
		return false
	}
	coheap.Remove(h, index)
	return false
}

type IntElem struct {
	data int
	*IndexMixin
}

func NewElem(data int) *IntElem {
	return &IntElem{
		data: data,
		IndexMixin: &IndexMixin{},
	}
}

func NewMaxHeap() *Heap {
	return MustHeap(func(i *IntElem, j *IntElem) bool {
		return i.data > j.data
	})
}

func NewMinHeap() *Heap {
	return MustHeap(func(i *IntElem, j *IntElem) bool {
		return i.data < j.data
	})
}
