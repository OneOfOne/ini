package ini

import "strconv"

// TODO optimize using an index rather than 3 strings
type KeyValue struct {
	key     string
	value   string
	comment string
}

func (kv *KeyValue) SetKey(v string) { kv.key = v }
func (kv KeyValue) Key() string      { return kv.key }

func (kv *KeyValue) SetValue(v string) { kv.value = v }
func (kv KeyValue) Value() string      { return kv.value }

func (kv *KeyValue) SetComment(v string) { kv.comment = v }
func (kv KeyValue) Comment() string      { return kv.comment }

func (kv *KeyValue) SetInt(v int64) { kv.value = strconv.FormatInt(v, 10) }

func (kv KeyValue) Int() (int64, error) {
	if kv.value == "" {
		return 0, ErrNoValue
	}
	base, s := 10, kv.value
	if len(s) > 2 {
		switch s[:2] {
		case "0x", "0X":
			s, base = s[2:], 16
		case "0b", "0B":
			s, base = s[2:], 2
		}
	}
	return strconv.ParseInt(s, base, 64)
}

func (kv KeyValue) IntDefault(v int64) int64 {
	if v, err := kv.Int(); err == nil {
		return v
	}
	return v
}

func (kv *KeyValue) SetUint(v uint64) { kv.value = strconv.FormatUint(v, 10) }

func (kv KeyValue) Uint() (uint64, error) {
	if kv.value == "" {
		return 0, ErrNoValue
	}
	base, s := 10, kv.value
	if len(s) > 2 {
		switch s[:2] {
		case "0x", "0X":
			s, base = s[2:], 16
		case "0b", "0B":
			s, base = s[2:], 2
		}
	}
	return strconv.ParseUint(s, base, 64)
}

func (kv KeyValue) UintDefault(v uint64) uint64 {
	if v, err := kv.Uint(); err == nil {
		return v
	}
	return v
}

func (kv *KeyValue) SetFloat(v float64) { kv.value = strconv.FormatFloat(v, 'E', -1, 64) }

func (kv KeyValue) Float() (float64, error) {
	if kv.value == "" {
		return 0, ErrNoValue
	}
	return strconv.ParseFloat(kv.value, 64)
}

func (kv KeyValue) FloatDefault(v float64) float64 {
	if v, err := kv.Float(); err == nil {
		return v
	}
	return v
}

func (kv *KeyValue) SetBool(v bool) { kv.value = strconv.FormatBool(v) }

func (kv KeyValue) Bool() (bool, error) {
	if kv.value == "" {
		return false, ErrNoValue
	}
	return strconv.ParseBool(kv.value)
}

func (kv KeyValue) BoolDefault(v bool) bool {
	if v, err := kv.Bool(); err == nil {
		return v
	}
	return v
}
