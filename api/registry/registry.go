package registry

import (
	"net"
	"sync"

	"github.com/glestaris/clique/api"
)

//go:generate counterfeiter . TransferStater
type TransferStater interface {
	TransferState() api.TransferState
}

type liveTransfer struct {
	spec       api.TransferSpec
	savedState api.TransferState
	stater     TransferStater
}

func (t *liveTransfer) state() api.TransferState {
	if t.savedState != api.TransferStateCompleted {
		t.savedState = t.stater.TransferState()
		if t.savedState == api.TransferStateCompleted {
			t.stater = nil
		}
	}

	return t.savedState
}

func (t *liveTransfer) transfer() api.Transfer {
	return api.Transfer{
		Spec:  t.spec,
		State: t.state(),
	}
}

type Registry struct {
	results    []api.TransferResults
	resultsMap map[string][]*api.TransferResults

	liveTransfers []liveTransfer

	lock sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		results:    make([]api.TransferResults, 0, 64),
		resultsMap: make(map[string][]*api.TransferResults),
	}
}

func (r *Registry) Transfers() []api.Transfer {
	r.lock.Lock()
	defer r.lock.Unlock()

	res := make([]api.Transfer, len(r.liveTransfers))
	for i, lt := range r.liveTransfers {
		res[i] = lt.transfer()
	}

	return res
}

func (r *Registry) TransfersByState(state api.TransferState) []api.Transfer {
	r.lock.Lock()
	defer r.lock.Unlock()

	res := []api.Transfer{}
	for _, lt := range r.liveTransfers {
		if lt.state() == state {
			res = append(res, lt.transfer())
		}
	}

	return res
}

func (r *Registry) RegisterTransfer(
	spec api.TransferSpec, stater TransferStater,
) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.liveTransfers = append(r.liveTransfers, liveTransfer{
		spec:   spec,
		stater: stater,
	})
}

func (r *Registry) TransferResults() []api.TransferResults {
	r.lock.Lock()
	defer r.lock.Unlock()

	res := make([]api.TransferResults, len(r.results))
	copy(res, r.results)

	return res
}

func (r *Registry) TransferResultsByIP(ip net.IP) []api.TransferResults {
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

func (r *Registry) RegisterResults(ip net.IP, res api.TransferResults) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.results = append(r.results, res)
	r.resultsMap[ip.String()] = append(r.resultsMap[ip.String()], &res)
}
