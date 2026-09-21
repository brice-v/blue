package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeModuleFixture(t *testing.T, dir, name, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestWildcardImportRemovesPrivateTopLevelNames(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "privmod.b", `val _hidden = "secret"
fun _helper() { "helper" }
fun public_fun() { _helper() }
`)

	c := New()
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from privmod import *")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}

	if _, ok := c.symbolTable.LookupDirect("_hidden"); ok {
		t.Error("private val _hidden leaked through a wildcard import")
	}
	if _, ok := c.symbolTable.LookupDirect("_helper"); ok {
		t.Error("private fun _helper leaked through a wildcard import")
	}
	if _, ok := c.symbolTable.LookupDirect("public_fun"); !ok {
		t.Error("public function public_fun was not imported")
	}
}

func TestWildcardImportRestoresShadowedPrivateName(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "shadowmod.b", `fun _shadowed() { "from module" }
`)

	c := New()
	existing := c.symbolTable.Define("_shadowed", true)
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from shadowmod import *")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}

	got, ok := c.symbolTable.LookupDirect("_shadowed")
	if !ok {
		t.Fatal("pre-existing private name was removed instead of restored")
	}
	if got.Index != existing.Index {
		t.Errorf("restored symbol index = %d, want %d", got.Index, existing.Index)
	}
}

func TestSelectiveImportOnlyCompilesRequestedDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "selmod.b", `val base = 10
fun helper(x) { x + base }
fun middle(x) { helper(x) * 2 }
fun requested() { middle(5) }
fun unused() { "unused" }
`)

	c := New()
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from selmod import {requested}")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}

	if _, ok := c.symbolTable.LookupDirect("requested"); !ok {
		t.Error("requested function was not imported")
	}
	if _, ok := c.symbolTable.LookupDirect("unused"); ok {
		t.Error("unused function was compiled by a selective import")
	}
	// Dependencies are pulled in transitively and stay under the module prefix.
	for _, dep := range []string{"selmod.helper", "selmod.middle", "selmod.base"} {
		if _, ok := c.symbolTable.LookupDirect(dep); !ok {
			t.Errorf("dependency %s was not compiled", dep)
		}
	}
	// A selective import exposes the requested names only, so no module
	// namespace is created.
	if _, ok := c.symbolTable.LookupDirect("selmod"); ok {
		t.Error("selective import should not define the module namespace")
	}
}

func TestSelectiveImportMissingDeclarationErrors(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "selmod.b", `fun present() { 1 }`)

	c := New()
	c.CompilerBasePath = dir
	err := c.Compile(parse("from selmod import {missing}"))
	if err == nil || !strings.Contains(err.Error(), "no such top level declaration") {
		t.Fatalf("expected missing declaration error, got %v", err)
	}
}

func TestSelectiveImportKeepsPrivateDependency(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "selmod.b", `fun _helper() { "helper" }
fun requested() { _helper() }
`)

	c := New()
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from selmod import {requested}")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}
	if _, ok := c.symbolTable.LookupDirect("selmod._helper"); !ok {
		t.Error("private dependency was not compiled")
	}
	if _, ok := c.symbolTable.LookupDirect("_helper"); ok {
		t.Error("private dependency leaked into the importing scope")
	}
}

func TestSelectiveImportScansModuleBodyStatements(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "selmod.b", `fun helper() { "helper" }
println(helper())
fun requested() { "requested" }
`)

	c := New()
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from selmod import {requested}")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}
	if _, ok := c.symbolTable.LookupDirect("selmod.helper"); !ok {
		t.Error("declaration used by a module body statement was not compiled")
	}
}

func TestSelectiveImportCompilesComprehensionDependency(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "selmod.b", `fun helper(x) { x + 1 }
fun requested() { [helper(i) for (i in 1..3)] }
`)

	c := New()
	c.CompilerBasePath = dir
	if err := c.Compile(parse("from selmod import {requested}")); err != nil {
		t.Fatalf("compile error: %s", err.Error())
	}
	if _, ok := c.symbolTable.LookupDirect("selmod.helper"); !ok {
		t.Error("declaration referenced from inside a comprehension was not compiled")
	}
}

func TestModulePrivateMemberAccessIsRejected(t *testing.T) {
	dir := t.TempDir()
	writeModuleFixture(t, dir, "privmod.b", `fun _hidden() { "hidden" }
fun public_fun() { _hidden() }
`)

	c := New()
	c.CompilerBasePath = dir
	err := c.Compile(parse("import privmod\nprivmod._hidden()"))
	if err == nil || !strings.Contains(err.Error(), "is private and cannot be accessed") {
		t.Fatalf("expected private member error, got %v", err)
	}
}
