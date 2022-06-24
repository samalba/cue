package adt_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"cuelang.org/go/cue/parser"
	"cuelang.org/go/internal/core/adt"
	"cuelang.org/go/internal/core/compile"
	"cuelang.org/go/internal/core/eval"
	"cuelang.org/go/internal/core/runtime"
	_ "cuelang.org/go/pkg"
)

func TestShowMe(t *testing.T) {
	code := `
a: b: "c"
a: d: e: "f"
a: g:  "h"
w: a
x: w.d.e
`
	f, err := parser.ParseFile("", code)
	if err != nil {
		t.Fatalf("could not parse: %v", err)
	}
	r := runtime.New()
	opCtx := eval.NewContext(r, nil)
	v, err := compile.Files(nil, opCtx, "", f)
	if err != nil {
		t.Fatalf("could not compile: %v", err)
	}
	v.Finalize(opCtx)
	g := newGrapher(opCtx)
	g.walk(reflect.ValueOf(v))
	fmt.Printf("%s", g.buf.Bytes())
}

func newGrapher(opCtx *adt.OpContext) *grapher {
	g := &grapher{
		visited: make(map[interface{}]bool),
		nodeIDs: make(map[interface{}]int),
		opCtx:   opCtx,
	}
	g.printf("flowchart TB\n")
	return g
}

type grapher struct {
	visited map[interface{}]bool
	nodeIDs map[interface{}]int
	buf     bytes.Buffer
	maxID   int
	opCtx   *adt.OpContext
}

func (g *grapher) printf(f string, a ...interface{}) {
	fmt.Fprintf(&g.buf, f, a...)
}

func (g *grapher) walk(x reflect.Value) {
	//	log.Printf("walk %v {", x.Type())
	//	defer log.Printf("}")
	x = bypassCanInterface(x)
	if x.Kind() != reflect.Ptr || x.IsNil() || x.Elem().Kind() != reflect.Struct {
		return
	}
	node := g.nodeID(x)
	if g.visited[x.Interface()] {
		return
	}
	g.visited[x.Interface()] = true
	g.printf("%s[\"%s\"]\n", node, g.nodeText(x))
	g.walkInner(node, "", x.Elem())
}

func (g *grapher) nodeText(xv reflect.Value) string {
	switch x := xv.Interface().(type) {
	case *adt.Field:
		return fmt.Sprintf("*adt.Field(%s)", x.Label.SelectorString(g.opCtx))
	case *adt.Vertex:
		return fmt.Sprintf("*adt.Vertex(%s)", x.Label.SelectorString(g.opCtx))
	case *adt.String:
		return fmt.Sprintf("*adt.String(%s)", x.Str)
	case *adt.SelectorExpr:
		return fmt.Sprintf("*adt.SelectorExpr(%s)", x.Sel.SelectorString(g.opCtx))
	case *adt.FieldReference:
		return fmt.Sprintf("*adt.FieldReference(%d, %s)", x.UpCount, x.Label.SelectorString(g.opCtx))
	default:
		return xv.Type().String()
	}
}

func (g *grapher) nodeID(v reflect.Value) string {
	id, ok := g.nodeIDs[v.Interface()]
	if !ok {
		id = g.maxID
		g.nodeIDs[v.Interface()] = id
		g.maxID++
	}
	return fmt.Sprintf("n%v", id)
}

func (g *grapher) walkInner(node string, arc string, v reflect.Value) {
	v = bypassCanInterface(v)
	//	log.Printf("walkInner %q %q %v{", node, arc, v.Type())
	//	defer log.Printf("}")
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			g.printf("%s -- \"%s\" --> %s\n", node, arc, g.nodeID(v))
			g.walk(v)
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			name := t.Field(i).Name
			if name == "Src" || name == "Structs" {
				continue
			}
			g.walkInner(node, arc+"."+name, v.Field(i))
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			g.walkInner(node, arc+"["+keyString(iter.Key())+"]", iter.Value())
		}
	case reflect.Interface:
		if !v.IsNil() {
			g.walkInner(node, arc, v.Elem())
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			g.walkInner(node, fmt.Sprintf("%s[%d]", arc, i), v.Index(i))
		}
	default:
		//		log.Printf("ignoring %v", v.Type())
	}
}

func keyString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

const (
	flagRO reflectFlag = 1<<5 | 1<<6
)

func interfaceOf(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}
	return bypassCanInterface(v).Interface()
}

type reflectFlag uintptr

var flagValOffset = func() uintptr {
	field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag")
	if !ok {
		panic("reflect.Value has no flag field")
	}
	return field.Offset
}()

func flagField(v *reflect.Value) *reflectFlag {
	return (*reflectFlag)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + flagValOffset))
}

// bypassCanInterface returns a version of v that
// bypasses the CanInterface check.
func bypassCanInterface(v reflect.Value) reflect.Value {
	if !v.IsValid() || v.CanInterface() {
		return v
	}
	*flagField(&v) &^= flagRO
	return v
}

// Sanity checks against future reflect package changes
// to the type or semantics of the Value.flag field.
func init() {
	field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag")
	if !ok {
		panic("reflect.Value has no flag field")
	}
	if field.Type.Kind() != reflect.TypeOf(reflectFlag(0)).Kind() {
		panic("reflect.Value flag field has changed kind")
	}
}
