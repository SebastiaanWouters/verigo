package object

func NewResultMap() *ResultMap {
	s := make(map[string]Object)
	return &ResultMap{store: s}
}

type ResultMap struct {
	store map[string]Object
}

func (e *ResultMap) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}
func (e *ResultMap) GetAll() map[string]Object {
	return e.store
}
func (e *ResultMap) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
