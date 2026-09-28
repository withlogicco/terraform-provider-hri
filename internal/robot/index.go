package robot

import "sync"

type Index struct {
	client  *Client
	once    sync.Once
	servers map[int]Server
	err     error
}

func NewIndex(c *Client) *Index { return &Index{client: c} }
func (i *Index) All() (map[int]Server, error) {
	i.once.Do(func() {
		var list []Server
		list, i.err = i.client.ListServers()
		if i.err != nil {
			return
		}
		i.servers = make(map[int]Server, len(list))
		for _, s := range list {
			i.servers[s.ServerNumber] = s
		}
	})
	return i.servers, i.err
}
func (i *Index) Get(n int) (Server, bool, error) {
	all, e := i.All()
	if e != nil {
		return Server{}, false, e
	}
	s, ok := all[n]
	return s, ok, nil
}
