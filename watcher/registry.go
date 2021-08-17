package watcher

import v1 "github.com/clambin/webmon/crds/targets/api/types/v1"

type registry struct {
	namespaces map[string]nameList
}

type nameList map[string]v1.TargetSpec

func newRegistry() *registry {
	return &registry{
		namespaces: make(map[string]nameList),
	}
}

func (r *registry) get(namespace, name string) (spec v1.TargetSpec, ok bool) {
	var names nameList
	names, ok = r.namespaces[namespace]

	if ok {
		spec, ok = names[name]
	}

	return
}

func (r *registry) add(namespace, name string, spec v1.TargetSpec) {
	_, ok := r.namespaces[namespace]

	if ok == false {
		r.namespaces[namespace] = make(nameList)
	}

	r.namespaces[namespace][name] = spec
}

func (r *registry) delete(namespace, name string) (spec v1.TargetSpec) {
	var ok bool
	spec, ok = r.get(namespace, name)

	if ok {
		delete(r.namespaces[namespace], name)
	}

	return
}
