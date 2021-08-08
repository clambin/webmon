package watcher

type registry struct {
	namespaces map[string]nameList
}

type nameList map[string]string

func newRegistry() *registry {
	return &registry{
		namespaces: make(map[string]nameList),
	}
}

func (r *registry) get(namespace, name string) (url string, ok bool) {
	var names nameList
	names, ok = r.namespaces[namespace]

	if ok {
		url, ok = names[name]
	}

	return
}

func (r *registry) add(namespace, name, url string) {
	_, ok := r.namespaces[namespace]

	if ok == false {
		r.namespaces[namespace] = make(nameList)
	}

	r.namespaces[namespace][name] = url
}

func (r *registry) delete(namespace, name string) (url string) {
	var ok bool
	url, ok = r.get(namespace, name)

	if ok {
		delete(r.namespaces[namespace], name)
	}

	return
}
