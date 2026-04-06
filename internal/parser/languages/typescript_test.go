package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zzet/gortex/internal/graph"
)

func TestTSExtractor_Function(t *testing.T) {
	src := []byte(`function greet(name: string): string {
  return "Hello " + name;
}
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("app.ts", src)
	require.NoError(t, err)

	funcs := nodesOfKind(result.Nodes, graph.KindFunction)
	require.Len(t, funcs, 1)
	assert.Equal(t, "greet", funcs[0].Name)
}

func TestTSExtractor_ArrowFunction(t *testing.T) {
	src := []byte(`const handler = () => {
  console.log("hello");
};
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("app.ts", src)
	require.NoError(t, err)

	funcs := nodesOfKind(result.Nodes, graph.KindFunction)
	require.Len(t, funcs, 1)
	assert.Equal(t, "handler", funcs[0].Name)
}

func TestTSExtractor_Class(t *testing.T) {
	src := []byte(`class UserService {
  getUser(id: string) {
    return {};
  }
}
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("service.ts", src)
	require.NoError(t, err)

	types := nodesOfKind(result.Nodes, graph.KindType)
	require.Len(t, types, 1)
	assert.Equal(t, "UserService", types[0].Name)

	methods := nodesOfKind(result.Nodes, graph.KindMethod)
	require.Len(t, methods, 1)
	assert.Equal(t, "getUser", methods[0].Name)
}

func TestTSExtractor_Interface(t *testing.T) {
	src := []byte(`interface Config {
  port: number;
  host: string;
}
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("types.ts", src)
	require.NoError(t, err)

	ifaces := nodesOfKind(result.Nodes, graph.KindInterface)
	require.Len(t, ifaces, 1)
	assert.Equal(t, "Config", ifaces[0].Name)
}

func TestTSExtractor_Variables(t *testing.T) {
	src := []byte(`const API_URL = "https://api.example.com";
let count = 0;
export const MAX_RETRIES = 3;
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("config.ts", src)
	require.NoError(t, err)

	vars := nodesOfKind(result.Nodes, graph.KindVariable)
	assert.GreaterOrEqual(t, len(vars), 2)
}

func TestTSExtractor_InterfaceMethods(t *testing.T) {
	src := []byte(`interface Repository {
    findById(id: string): User;
    save(user: User): void;
}
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("repo.ts", src)
	require.NoError(t, err)

	ifaces := nodesOfKind(result.Nodes, graph.KindInterface)
	require.Len(t, ifaces, 1)
	methods, ok := ifaces[0].Meta["methods"].([]string)
	require.True(t, ok)
	assert.Contains(t, methods, "findById")
	assert.Contains(t, methods, "save")
}

func TestTSExtractor_Imports(t *testing.T) {
	src := []byte(`import { Router } from 'express';
import axios from 'axios';
`)
	e := NewTypeScriptExtractor()
	result, err := e.Extract("app.ts", src)
	require.NoError(t, err)

	imports := edgesOfKind(result.Edges, graph.EdgeImports)
	require.Len(t, imports, 2)
}
