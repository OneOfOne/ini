package ini

import (
	"log"
	"os"
	"strings"
	"testing"
)

const exFile = `
top = 1

[core]
	top = ${top} // comment
	core 1 top with dot = ${core.1.top}
	core 1 top with space  = ${core.1 top}

[core.1]
	multiLineValue = ''' multi
	line
	value
'''
	top = ${core.top}

[core.other] // comment
	xx = true
	noValue # comment

[boo]
	xxx = 1

[core]
	top = override

[core.copy]
	%inc(default)

[core.copy.copy]
	%inc(core.copy)

[core.copy.copy.copy]
	%inc(core.copy.copy)
	a = override

[default]
	a = 1
	b = 2
	c = 3

`

func init() {
	log.SetFlags(log.Lshortfile)
}

func TestReadWrite(t *testing.T) {
	// n := Parse(strings.NewReader(exFile))
	// var buf bytes.Buffer
	// if _, err := n.WriteTo(&buf); err != nil {
	// 	t.Fatal(err)
	// }
	// if exFile != buf.String() {
	// 	t.Fatalf("exFile != buf.String()\n%s", buf.Bytes())
	// }
	// t.Log(n.FindValue("core.int.xz", "XX", true))
	// t.Log(n.FindValue("", "TOP", false))
	var ss Sections
	ss.ReadFrom(strings.NewReader(exFile))
	ss.expand()
	// ss = ss.ExpandAll()
	// j, _ := json.MarshalIndent(ss, "", "  ")
	// t.Logf("%s", j)
	// for _, s := range ss.ss {
	// 	for _, kv := range s.kvs {
	// 		t.Logf("%s.%s = %s", s.name, kv.key, kv.value)
	// 	}
	// }
	ss.WriteTo(os.Stderr)
	// t.Log(ss.Section("core").ValueByKey("test").Int())
	// t.Log(ss.ExpandByKey("core", "test").Int())
	// t.Log(ss.ExpandByKey("core.int", "yy").Int())
}
