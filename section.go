package ini

import "sort"

type Section struct {
	name    string
	comment string
	kvs     []KeyValue
	idx     uint16
}

func (s *Section) Len() int { return len(s.kvs) }

func (s *Section) IndexOf(key string) int {
	ln := s.Len()
	i := sort.Search(ln, func(i int) bool { return s.kvs[i].key >= key })
	if i == ln || s.kvs[i].key != key {
		return -1
	}
	return i
}

func (s *Section) Get(key string) *KeyValue {
	if i := s.IndexOf(key); i != -1 {
		return &s.kvs[i]
	}
	return nil
}

func (s *Section) GetOrCreate(key string) *KeyValue {
	if kv := s.Get(key); kv != nil {
		return kv
	}
	s.kvs = append(s.kvs, KeyValue{key: key})
	sort.Sort(kvSortByKey{s.kvs})
	return &s.kvs[s.IndexOf(key)]
}

func (s *Section) set(okv KeyValue) *KeyValue {
	kv := s.GetOrCreate(okv.key)
	*kv = okv
	return kv
}
