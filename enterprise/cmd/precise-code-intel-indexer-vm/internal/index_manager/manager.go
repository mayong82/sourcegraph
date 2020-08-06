package indexmanager

import "sync"

type Manager struct {
	m        sync.RWMutex
	indexIDs map[int]struct{}
}

func New() *Manager {
	return &Manager{
		indexIDs: map[int]struct{}{},
	}
}

func (i *Manager) GetIDs() (ids []int) {
	i.m.RLock()
	defer i.m.RUnlock()

	for id := range i.indexIDs {
		ids = append(ids, id)
	}

	return ids
}

func (i *Manager) AddID(indexID int) {
	i.m.Lock()
	i.indexIDs[indexID] = struct{}{}
	i.m.Unlock()
}

func (i *Manager) RemoveID(indexID int) {
	i.m.Lock()
	delete(i.indexIDs, indexID)
	i.m.Unlock()
}
