package main

import (
	"fmt"
	"reflect"
	"sort"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/internal/core/adt"
)

func main() {
	allTypes := allASTTypes
	sort.Slice(allTypes, func(i, j int) bool {
		t1, t2 := allTypes[i], allTypes[j]
		if (t1.Kind() == reflect.Interface) == (t2.Kind() == reflect.Interface) {
			return t1.Name() < t2.Name()
		}
		return t1.Kind() == reflect.Interface
	})
	fmt.Printf("implements")
	for _, t := range allTypes {
		if t.Kind() == reflect.Interface {
			fmt.Printf(",%s", t.Name())
		}
	}
	fmt.Println()
	for _, t := range allTypes {
		fmt.Printf("%s", t.Name())
		for _, it := range allTypes {
			if it.Kind() != reflect.Interface {
				continue
			}
			switch {
			case t.Implements(it):
				fmt.Printf(",y")
			case reflect.PtrTo(t).Implements(it):
				fmt.Printf(",*y")
			default:
				fmt.Printf(",")
			}
		}
		fmt.Println()
	}
}

func typeOf(x any) reflect.Type {
	return reflect.TypeOf(x).Elem()
}

var allADTTypes = []reflect.Type{
	typeOf(new(adt.Node)),
	typeOf(new(adt.Decl)),
	typeOf(new(adt.Elem)),
	typeOf(new(adt.Expr)),
	typeOf(new(adt.BaseValue)),
	typeOf(new(adt.Value)),
	typeOf(new(adt.Evaluator)),
	typeOf(new(adt.Resolver)),
	typeOf(new(adt.YieldFunc)),
	typeOf(new(adt.Yielder)),
	typeOf(new(adt.Validator)),
	typeOf(new(adt.CloseInfo)),
	typeOf(new(adt.SpanType)),
	typeOf(new(adt.Environment)),
	typeOf(new(adt.ID)),
	typeOf(new(adt.Vertex)),
	typeOf(new(adt.StructInfo)),
	typeOf(new(adt.VertexStatus)),
	typeOf(new(adt.OptionalType)),
	typeOf(new(adt.Conjunct)),
	typeOf(new(adt.Runtime)),
	typeOf(new(adt.Config)),
	typeOf(new(adt.OpContext)),
	typeOf(new(adt.Flag)),
	typeOf(new(adt.ErrorCode)),
	typeOf(new(adt.Bottom)),
	typeOf(new(adt.ValueError)),
	typeOf(new(adt.Stats)),
	typeOf(new(adt.StructLit)),
	typeOf(new(adt.FieldInfo)),
	typeOf(new(adt.Field)),
	typeOf(new(adt.OptionalField)),
	typeOf(new(adt.BulkOptionalField)),
	typeOf(new(adt.Ellipsis)),
	typeOf(new(adt.DynamicField)),
	typeOf(new(adt.ListLit)),
	typeOf(new(adt.Null)),
	typeOf(new(adt.Bool)),
	typeOf(new(adt.Num)),
	typeOf(new(adt.String)),
	typeOf(new(adt.Bytes)),
	typeOf(new(adt.ListMarker)),
	typeOf(new(adt.StructMarker)),
	typeOf(new(adt.Top)),
	typeOf(new(adt.BasicType)),
	typeOf(new(adt.BoundExpr)),
	typeOf(new(adt.BoundValue)),
	typeOf(new(adt.NodeLink)),
	typeOf(new(adt.FieldReference)),
	typeOf(new(adt.ValueReference)),
	typeOf(new(adt.LabelReference)),
	typeOf(new(adt.DynamicReference)),
	typeOf(new(adt.ImportReference)),
	typeOf(new(adt.LetReference)),
	typeOf(new(adt.SelectorExpr)),
	typeOf(new(adt.IndexExpr)),
	typeOf(new(adt.SliceExpr)),
	typeOf(new(adt.Interpolation)),
	typeOf(new(adt.UnaryExpr)),
	typeOf(new(adt.BinaryExpr)),
	typeOf(new(adt.CallExpr)),
	typeOf(new(adt.Builtin)),
	typeOf(new(adt.Param)),
	typeOf(new(adt.BuiltinValidator)),
	typeOf(new(adt.DisjunctionExpr)),
	typeOf(new(adt.Disjunct)),
	typeOf(new(adt.Conjunction)),
	typeOf(new(adt.Disjunction)),
	typeOf(new(adt.Comprehension)),
	typeOf(new(adt.ForClause)),
	typeOf(new(adt.IfClause)),
	typeOf(new(adt.LetClause)),
	typeOf(new(adt.ValueClause)),
	typeOf(new(adt.Feature)),
	typeOf(new(adt.StringIndexer)),
	typeOf(new(adt.FeatureType)),
	typeOf(new(adt.Concreteness)),
	typeOf(new(adt.Kind)),
}

var allASTTypes = []reflect.Type{
	typeOf(new(ast.Node)),
	typeOf(new(ast.Expr)),
	typeOf(new(ast.Decl)),
	typeOf(new(ast.Label)),
	typeOf(new(ast.Clause)),
	typeOf(new(ast.Comment)),
	typeOf(new(ast.CommentGroup)),
	typeOf(new(ast.Attribute)),
	typeOf(new(ast.Field)),
	typeOf(new(ast.Alias)),
	typeOf(new(ast.Comprehension)),
	typeOf(new(ast.BadExpr)),
	typeOf(new(ast.BottomLit)),
	typeOf(new(ast.Ident)),
	typeOf(new(ast.BasicLit)),
	typeOf(new(ast.Interpolation)),
	typeOf(new(ast.StructLit)),
	typeOf(new(ast.ListLit)),
	typeOf(new(ast.Ellipsis)),
	typeOf(new(ast.ForClause)),
	typeOf(new(ast.IfClause)),
	typeOf(new(ast.LetClause)),
	typeOf(new(ast.ParenExpr)),
	typeOf(new(ast.SelectorExpr)),
	typeOf(new(ast.IndexExpr)),
	typeOf(new(ast.SliceExpr)),
	typeOf(new(ast.CallExpr)),
	typeOf(new(ast.UnaryExpr)),
	typeOf(new(ast.BinaryExpr)),
	typeOf(new(ast.ImportSpec)),
	typeOf(new(ast.BadDecl)),
	typeOf(new(ast.ImportDecl)),
	typeOf(new(ast.Spec)),
	typeOf(new(ast.EmbedDecl)),
	typeOf(new(ast.File)),
	typeOf(new(ast.Package)),
}
