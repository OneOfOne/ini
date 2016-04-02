package ini

import (
	"errors"
	"fmt"
)

const (
	multiLineChars = `'''`
	deleteIdx      = ^uint16(0)
)

var (
	ErrNoValue = errors.New("empty value")
)

type IncludeError struct {
	Section, Inc string
}

func (ie IncludeError) Error() string {
	return fmt.Sprintf("<bad include %q in section %q>", ie.Inc, ie.Section)
}

type ValueError struct {
	Section, Key, Exp string
}

func (ve ValueError) Error() string {
	return fmt.Sprintf("<bad expand ${%s} in section %q, key %q>", ve.Exp, ve.Section, ve.Key)
}

type kvSortByKey struct{ kv []KeyValue }

func (s kvSortByKey) Len() int      { return len(s.kv) }
func (s kvSortByKey) Swap(i, j int) { s.kv[i], s.kv[j] = s.kv[j], s.kv[i] }
func (s kvSortByKey) Less(i, j int) bool {
	if s.kv[i].key == "" {
		return false
	}
	if s.kv[j].key == "" {
		return true
	}
	return s.kv[i].key < s.kv[j].key
}

type secSortByName struct{ ss []Section }

func (s secSortByName) Len() int           { return len(s.ss) }
func (s secSortByName) Swap(i, j int)      { s.ss[i], s.ss[j] = s.ss[j], s.ss[i] }
func (s secSortByName) Less(i, j int) bool { return s.ss[i].name < s.ss[j].name }

type secSortByIdx struct{ secSortByName }

func (s secSortByIdx) Less(i, j int) bool { return s.ss[i].idx < s.ss[j].idx }
