// Copyright The OpenTelemetry Authors
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

package lib

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Analysis struct {
	ProjectPath    string
	PackagePattern string
	RootFunctions  []FuncDescriptor
	FuncDecls      map[FuncDescriptor]bool
	Callgraph      map[FuncDescriptor][]FuncDescriptor
	Interfaces     map[string]bool
}

type Import struct {
	NamedPackage string
	Package      string
}

type AnalysisPass interface {
	Execute(node *ast.File,
		analysis *Analysis,
		pkg *packages.Package) []Import
}

func (analysis *Analysis) Execute(pass AnalysisPass, inputFileSuffix string, intermediateFileSuffix string) {
	fset := token.NewFileSet()
	fmt.Println("Instrumentation")
	cfg := &packages.Config{Fset: fset, Mode: LoadMode, Dir: analysis.ProjectPath}
	pkgs, err := packages.Load(cfg, analysis.PackagePattern)
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		var node *ast.File
		for _, node = range pkg.Syntax {
			var out *os.File
			fmt.Println("\t\t", fset.File(node.Pos()).Name())
			if len(inputFileSuffix) > 0 {
				out, _ = os.Create(fset.File(node.Pos()).Name() + inputFileSuffix)
				defer out.Close()
			} else {
				out, _ = os.Create(fset.File(node.Pos()).Name() + intermediateFileSuffix)
				defer out.Close()
			}
			if len(analysis.RootFunctions) == 0 {
				printer.Fprint(out, fset, node)
				continue
			}
			imports := pass.Execute(node, analysis, pkg)
			for _, imp := range imports {
				if len(imp.NamedPackage) > 0 {
					astutil.AddNamedImport(fset, node, imp.NamedPackage, imp.Package)
				} else {
					astutil.AddImport(fset, node, imp.Package)
				}
			}
			printer.Fprint(out, fset, node)
			if len(inputFileSuffix) > 0 {
				os.Rename(fset.File(node.Pos()).Name(), fset.File(node.Pos()).Name()+".original")
			} else {
				os.Rename(fset.File(node.Pos()).Name()+intermediateFileSuffix, fset.File(node.Pos()).Name())
			}

		}
	}
}
