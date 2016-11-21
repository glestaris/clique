package testhelpers

import "math/rand"

var allocatedPorts set

func init() {
	allocatedPorts = set{
		data: make(map[uint16]bool),
	}
}

type set struct {
	data map[uint16]bool
}

func (s set) add(i uint16) {
	s.data[i] = true
}

func (s set) has(i uint16) bool {
	_, ok := s.data[i]
	return ok
}

func (s set) remove(i uint16) {
	delete(s.data, i)
}

func SelectPort(node int) uint16 {
	for {
		port := uint16((1000 * node) + rand.Intn(1000))
		if allocatedPorts.has(port) {
			continue
		}

		allocatedPorts.add(port)
		return port
	}
}

func ReleasePort(port uint16) {
	allocatedPorts.remove(port)
}
