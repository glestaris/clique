package registry

import (
	"net"
	"sync"

	"github.com/glestaris/ice-clique/api"
)

type Registry struct {
	results    []api.TransferResults
	resultsMap map[string][]*api.TransferResults

	lock sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		results:    make([]api.TransferResults, 0, 64),
		resultsMap: make(map[string][]*api.TransferResults),
	}
}

func (r *Registry) Transfers() []api.TransferResults {
	r.lock.Lock()
	defer r.lock.Unlock()

	res := make([]api.TransferResults, len(r.results))
	copy(res, r.results)

	return res
}

func (r *Registry) TransfersByIP(ip net.IP) []api.TransferResults {
	r.lock.Lock()
	defer r.lock.Unlock()

	resPtrs := r.resultsMap[ip.String()]
	if len(resPtrs) == 0 {
		return []api.TransferResults{}
	}

	res := make([]api.TransferResults, len(resPtrs))
	for i := 0; i < len(resPtrs); i++ {
		res[i] = *resPtrs[i]
	}

	return res
}

func (r *Registry) Register(ip net.IP, res api.TransferResults) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.results = append(r.results, res)
	r.resultsMap[ip.String()] = append(r.resultsMap[ip.String()], &res)
}
