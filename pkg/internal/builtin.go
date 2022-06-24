// Copyright 2020 CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/internal"
	"cuelang.org/go/internal/core/adt"
	"cuelang.org/go/internal/core/compile"
	"cuelang.org/go/internal/core/convert"
	"cuelang.org/go/internal/core/runtime"
)

// A Builtin is a Builtin function or constant.
//
// A function may return and a constant may be any of the following types:
//
//	error (translates to bottom)
//	nil   (translates to null)
//	bool
//	int*
//	uint*
//	float64
//	string
//	*big.Float
//	*big.Int
//
//	For any of the above, including interface{} and these types recursively:
//	[]T
//	map[string]T
type XBuiltin struct {
	Name   string
	Pkg    adt.Feature
	Params []Param
	Result adt.Kind
	Func   func(c *CallCtxt)
	Const  string
}

type Param struct {
	Kind  adt.Kind
	Value adt.Value // input constraint (may be nil)
}

type Package struct {
	Funcs map[string]func(c *CallCtxt)
	CUE   string
}

func newContext() *cue.Context {
	return (*cue.Context)(runtime.New())
}

var (
	funcPath  = cue.MakePath(cue.Str("func"))
	valuePath = cue.MakePath(cue.Str("value"))
)

func (p *Package) MustCompile(opCtx *adt.OpContext, importPath string) *adt.Vertex {
	v, err := p.compile(opCtx, importPath)
	if err != nil {
		panic(fmt.Errorf("cannot compile builtin package %q: %v", importPath, err))
	}
	return v
}

func (p *Package) compile(opCtx *adt.OpContext, importPath string) (*adt.Vertex, error) {
	f, err := parser.ParseFile(importPath, p.CUE)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q: %v", p.CUE, err)
	}
	pv, err := compile.Files(nil, opCtx, importPath, f)
	if err != nil {
		return nil, fmt.Errorf("could not compile %q: %v", p.CUE, err)
	}
	pv.Finalize(opCtx)
	if k := pv.Kind(); k != adt.StructKind {
		return nil, fmt.Errorf("top level is %v not struct", k)
	}

	c := &compiler{
		opCtx:           opCtx,
		pkg:             p,
		importPath:      importPath,
		importPathLabel: opCtx.StringLabel(importPath),
		inLabel:         opCtx.StringLabel("in"),
		outLabel:        opCtx.StringLabel("out"),
	}
	funcsLabel := opCtx.StringLabel("funcs")
	newArcs := make([]*adt.Vertex, 0, len(pv.Arcs))
	for _, a := range pv.Arcs {
		if a.Label != funcsLabel {
			// It's a straight-up value: just add it as-is.
			newArcs = append(newArcs, a)
			continue
		}
		// It's the funcs struct. Add each member as a built-in function.
		for _, funcArc := range a.Arcs {
			b, err := c.newBuiltin(funcArc)
			if err != nil {
				return nil, fmt.Errorf("cannot make builtin for %s: %v", funcArc.Label.StringValue(opCtx), err)
			}
			// NB this is pretty sleazy.
			funcArc.BaseValue = b
			funcArc.Conjuncts = nil
			funcArc.Structs = nil
			newArcs = append(newArcs, funcArc)
		}
	}
	pv.Arcs = newArcs

	return pv, nil
}

// compiler is a utility type that allows us to avoid recomputing
// labels for each builtin function.
type compiler struct {
	pkg               *Package
	opCtx             *adt.OpContext
	importPath        string
	importPathLabel   adt.Feature
	inLabel, outLabel adt.Feature // "in", "out"
}

func (c *compiler) newBuiltin(bv *adt.Vertex) (*adt.Builtin, error) {
	if k := bv.Kind(); k != adt.StructKind {
		return nil, fmt.Errorf("builtin info is %v not struct", k)
	}
	params := make([]adt.Param, 0, 3)
	in := bv.Lookup(c.inLabel)
	if in == nil {
		return nil, fmt.Errorf("no in parameter found")
	}
	if k := in.Kind(); k != adt.StructKind {
		return nil, fmt.Errorf("in parameter is %v not struct", k)
	}
	for _, arg := range in.Arcs {
		argName := arg.Label.IdentString(c.opCtx)
		if !strings.HasPrefix(argName, "#A") {
			// TODO allow named arguments instead of panicking.
			return nil, fmt.Errorf("unexpected arg field %q", argName)
		}
		argIdx, err := strconv.Atoi(argName[2:])
		if err != nil {
			return nil, fmt.Errorf("invalid arg index in arg %q", argName)
		}
		if argIdx < 0 {
			return nil, fmt.Errorf("negative arg index in arg %q", argName)
		}
		if argIdx >= len(params) {
			for j := len(params); j <= argIdx; j++ {
				params = append(params, adt.Param{})
			}
		}
		params[argIdx] = adt.Param{
			Value: arg,
		}
	}
	for i, p := range params {
		if p.Value == nil {
			return nil, fmt.Errorf("no parameter type defined for argument %d", i)
		}
	}
	out := bv.Lookup(c.outLabel)
	if out == nil {
		return nil, fmt.Errorf("no out parameter found")
	}
	funcNameLabel := bv.Label
	funcName := funcNameLabel.StringValue(c.opCtx)
	f := c.pkg.Funcs[funcName]
	if f == nil {
		return nil, fmt.Errorf("no underlying function found")
	}
	return &adt.Builtin{
		Package: c.importPathLabel,
		Name:    funcName,
		Func:    c.makeFunc(f, funcNameLabel),
		Params:  params,
		Result:  out,
	}, nil
}

func (c *compiler) makeFunc(f func(c *CallCtxt), funcName adt.Feature) func(ctx *adt.OpContext, args []adt.Value) adt.Expr {
	return func(ctx *adt.OpContext, args []adt.Value) (ret adt.Expr) {
		callCtx := &CallCtxt{
			ctx:        ctx,
			args:       args,
			importPath: c.importPathLabel,
			funcName:   funcName,
		}
		log.Printf("calling %v", callCtx.Name())
		defer func() {
			var errVal interface{} = callCtx.Err
			if err := recover(); err != nil {
				errVal = err
			}
			ret = processErr(callCtx, errVal, ret)
		}()
		f(callCtx)
		switch v := callCtx.Ret.(type) {
		case nil:
			// Validators may return a nil in case validation passes.
			return nil
		case *adt.Bottom:
			// deal with API limitation: catch nil interface issue.
			if v != nil {
				return v
			}
			return nil
		case adt.Value:
			return v
		case bottomer:
			// deal with API limitation: catch nil interface issue.
			if b := v.Bottom(); b != nil {
				return b
			}
			return nil
		}
		if callCtx.Err != nil {
			return nil
		}
		return convert.GoValueToValue(ctx, callCtx.Ret, true)
	}
}

func processErr(call *CallCtxt, errVal interface{}, ret adt.Expr) adt.Expr {
	ctx := call.ctx
	switch err := errVal.(type) {
	case nil:
	case *adt.Bottom:
		ret = err
	case *callError:
		ret = err.b
	case *json.MarshalerError:
		if err, ok := err.Err.(bottomer); ok {
			if b := err.Bottom(); b != nil {
				ret = b
			}
		}
	case bottomer:
		ret = wrapCallErr(call, err.Bottom())

	case errors.Error:
		// Convert lists of errors to a combined Bottom error.
		if list := errors.Errors(err); len(list) != 0 && list[0] != errVal {
			var errs *adt.Bottom
			for _, err := range list {
				if b, ok := processErr(call, err, nil).(*adt.Bottom); ok {
					errs = adt.CombineErrors(nil, errs, b)
				}
			}
			if errs != nil {
				return errs
			}
		}

		ret = wrapCallErr(call, &adt.Bottom{Err: err})
	case error:
		if call.Err == internal.ErrIncomplete {
			err := ctx.NewErrf("incomplete value")
			err.Code = adt.IncompleteError
			ret = err
		} else {
			// TODO: store the underlying error explicitly
			ret = wrapCallErr(call, &adt.Bottom{Err: errors.Promote(err, "")})
		}
	default:
		// Likely a string passed to panic.
		ret = wrapCallErr(call, &adt.Bottom{
			Err: errors.Newf(call.Pos(), "%s", err),
		})
	}
	return ret
}
