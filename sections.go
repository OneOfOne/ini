package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Sections struct {
	ss  []*Section
	idx uint16
}

func (ss *Sections) Len() int { return len(ss.ss) }

func (ss *Sections) IndexOf(name string) int {
	ln := ss.Len()
	i := sort.Search(ln, func(i int) bool { return ss.ss[i].name >= name })
	if i == ln || ss.ss[i].name != name {
		return -1
	}
	return i
}

func (ss *Sections) Get(name string) *Section {
	if i := ss.IndexOf(name); i != -1 {
		return ss.ss[i]
	}
	return nil
}

func (ss *Sections) GetOrCreate(name, comment string) *Section {
	if s := ss.Get(name); s != nil {
		if comment != "" {
			if s.comment == "" {
				s.comment = comment
			} else {
				s.comment += "\n" + comment
			}
		}
		return s
	}
	ss.ss = append(ss.ss, &Section{name: name, comment: comment, idx: ss.idx})
	ss.idx++
	sort.Sort(secSortByName{ss.ss})
	return ss.ss[ss.IndexOf(name)]
}

var expandRe = regexp.MustCompile(`\$\{[^}]+\}`)

func (ss *Sections) expand() (err error) {
	re, pass := expandRe.Copy(), 0
	for {
		pass++
		recheck := false
		for i := range ss.ss {
			s := ss.ss[i]
		REDO:
			for i := range s.kvs {
				kv := &s.kvs[i]
				if strings.HasPrefix(kv.key, "%inc(") && kv.key[len(kv.key)-1] == ')' {
					kw := kv.key[5 : len(kv.key)-1]
					oSec := ss.Get(kw)
					if oSec == nil || oSec == s {
						return IncludeError{s.name, kw}
					}
					for _, okv := range oSec.kvs {
						if s.IndexOf(okv.key) == -1 {
							s.set(okv)
							recheck = true
						}
					}
					*kv = KeyValue{}
					trimKVs(s)
					goto REDO
				}

				kv.value = re.ReplaceAllStringFunc(kv.value, func(kw string) string {
					if err != nil {
						return ""
					}
					var sec string
					kw = kw[2 : len(kw)-1] // strip ${}

					if idx := strings.LastIndexAny(kw, " ."); idx > -1 {
						sec, kw = kw[:idx], kw[idx+1:]
					}
					if sec == s.name && kv.key == kw {
						err = ValueError{s.name, kv.key, strings.TrimLeft(sec+" "+kw, " ")}
						return ""
					}
					if s := ss.Get(sec); s != nil {
						if okv := s.Get(kw); okv != nil {
							return okv.value
						}
					}
					err = ValueError{s.name, kv.key, strings.TrimLeft(sec+" "+kw, " ")}
					return ""
				})

				if err != nil {
					return
				}

				if re.MatchString(kv.value) {
					recheck = true
				}
			}
		}
		if !recheck {
			break
		}

		if pass > 2 {
			panic("this should never happen")
		}
	}
	return
}

func trimKVs(s *Section) {
	sort.Sort(kvSortByKey{s.kvs})
	i := len(s.kvs) - 1
	for ; i > -1 && s.kvs[i].key == ""; i-- {
	}
	s.kvs = s.kvs[:i+1 : i+1]
}

func (ss *Sections) ReadFrom(r io.Reader) (total int64, err error) {
	var (
		sc  = bufio.NewScanner(r)
		cur = ss.GetOrCreate("", "")

		ml     uint8
		kv     KeyValue
		expand bool
	)
	for sc.Scan() {
		s := sc.Text()
		ln := len(s)
		total += int64(ln) + 1

		if ml != 0 {
			mlIdx := strings.Index(s, multiLineChars)
			if mlIdx != -1 {
				s = s[:mlIdx]
			}
			switch ml {
			case 1:
				kv.key += "\n" + s
			case 2:
				kv.value += "\n" + s
			}
			if mlIdx != -1 {
				ml = 0
				cur.set(kv)
			}
			continue
		}
		s = strings.TrimSpace(s)

		if ln == 0 {
			continue
		}

		if !expand && strings.IndexAny(s, "$%") != -1 {
			expand = true
		}

		if name, comment, ok := getSecComment(s); ok {
			cur = ss.GetOrCreate(name, comment)
			continue
		}
		if kv, ml = getKeyValue(s); ml == 0 {
			cur.set(kv)
		}
	}
	if expand {
		err = ss.expand()
	}
	return
}

func (ss *Sections) WriteTo(w io.Writer) (total int64, err error) {
	var n int
	for i := range ss.ss {
		s := ss.ss[i]
		if s.Len() == 0 {
			continue
		}
		if s.name != "" {
			if s.comment != "" {
				n, err = fmt.Fprintf(w, "[%s] // %s\n", s.name, s.comment)
			} else {
				n, err = fmt.Fprintf(w, "[%s]\n", s.name)
			}
			if total += int64(n); err != nil {
				return
			}
		}
		for i := range s.kvs {
			kv := &s.kvs[i]
			if k := kv.Key(); k != "" {
				if s.name == "" {
					n, err = fmt.Fprintf(w, "%s", k)
				} else {
					n, err = fmt.Fprintf(w, "\t%s", k)
				}
				total += int64(n)
			}
			if v := kv.Value(); v != "" {
				if strings.Contains(v, "\n") {
					n, err = fmt.Fprintf(w, " = %s%s%s", multiLineChars, v, multiLineChars)
				} else {
					n, err = fmt.Fprintf(w, " = %s", v)
				}
				total += int64(n)
			}
			if c := kv.Comment(); c != "" {
				if kv.Key() == "" {
					n, err = fmt.Fprintf(w, "\t// %s", c)
				} else {
					n, err = fmt.Fprintf(w, " // %s", c)
				}
				total += int64(n)
			}
			if _, err = w.Write([]byte("\n")); err != nil {
				return
			}
			total++
		}
		if i == len(ss.ss)-1 {
			continue
		}
		if _, err = w.Write([]byte("\n")); err != nil {
			return
		}
	}
	return
}

func (ss Sections) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i := range ss.ss {
		s := ss.ss[i]
		if s.name != "" {
			fmt.Fprintf(&buf, "%q:{", s.name)
		}
		for i := range s.kvs {
			v := &s.kvs[i]

			if v.key == "" {
				continue
			}
			fmt.Fprintf(&buf, "%q: %q,", v.key, v.value)
		}
		if buf.Bytes()[buf.Len()-1] == ',' {
			buf.Truncate(buf.Len() - 1)
		}
		if s.name != "" {
			buf.WriteString("},")
		} else {
			buf.WriteString(",")
		}

	}
	if buf.Bytes()[buf.Len()-1] == ',' {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func getSecComment(s string) (name, comment string, ok bool) {
	if s[0] != '[' {
		return
	}
	var last byte
	for i := range s {
		if i == 0 {
			continue
		}
		switch s[i] {
		case ']':
			name, ok = strings.TrimSpace(s[1:i]), true
		case '/':
			if ok && last == '/' {
				comment = strings.TrimLeftFunc(s[i+1:], unicode.IsSpace)
			}
		case '#':
			if ok {
				comment = strings.TrimLeftFunc(s[i+1:], unicode.IsSpace)
			}

		}
		last = s[i]
	}
	return
}

func getKeyValue(s string) (kv KeyValue, ml uint8) {
	var (
		cIdx, eqIdx = len(s), -1
		maybe       bool
		last        byte
	)
L:
	for i := range s {
		switch s[i] {
		case '=':
			if eqIdx == -1 {
				kv.key, eqIdx = strings.TrimRightFunc(s[:i], unicode.IsSpace), i+1
			}
		case ' ':
			maybe = true
		case '/':
			if (maybe || i == 1) && last == '/' {
				cIdx = i - 1
				kv.comment = s[i+1 : len(s)]
				break L
			}
		case '#':
			if maybe || i == 0 {
				cIdx = i
				kv.comment = s[i+1 : len(s)]
				break L
			}
		default:
			maybe = i == 0
		}
		last = s[i]
	}
	if eqIdx != -1 {
		if kv.value = strings.TrimSpace(s[eqIdx:cIdx]); strings.HasPrefix(kv.value, multiLineChars) {
			kv.value, ml = kv.value[3:], 2
		}
	} else {
		if kv.key = strings.TrimRightFunc(s[:cIdx], unicode.IsSpace); strings.HasPrefix(kv.key, multiLineChars) {
			kv.key, ml = kv.key[3:], 1
		}
	}
	kv.comment = strings.TrimLeftFunc(kv.comment, unicode.IsSpace)
	return
}
