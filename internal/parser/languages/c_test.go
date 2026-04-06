package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zzet/gortex/internal/graph"
)

func TestCExtractor_Function(t *testing.T) {
	src := []byte(`#include <stdio.h>

void greet(const char* name) {
    printf("Hello %s\n", name);
}

int add(int a, int b) {
    return a + b;
}
`)
	e := NewCExtractor()
	result, err := e.Extract("main.c", src)
	require.NoError(t, err)

	funcs := nodesOfKind(result.Nodes, graph.KindFunction)
	assert.Len(t, funcs, 2)
}

func TestCExtractor_Struct(t *testing.T) {
	src := []byte(`struct Point {
    int x;
    int y;
};
`)
	e := NewCExtractor()
	result, err := e.Extract("point.c", src)
	require.NoError(t, err)

	types := nodesOfKind(result.Nodes, graph.KindType)
	assert.GreaterOrEqual(t, len(types), 1)
}

func TestCExtractor_Include(t *testing.T) {
	src := []byte(`#include <stdio.h>
#include "mylib.h"
`)
	e := NewCExtractor()
	result, err := e.Extract("main.c", src)
	require.NoError(t, err)

	imports := edgesOfKind(result.Edges, graph.EdgeImports)
	assert.Len(t, imports, 2)
}

func TestCExtractor_Enum(t *testing.T) {
	src := []byte(`enum Color {
    RED,
    GREEN,
    BLUE
};
`)
	e := NewCExtractor()
	result, err := e.Extract("color.c", src)
	require.NoError(t, err)

	types := nodesOfKind(result.Nodes, graph.KindType)
	require.GreaterOrEqual(t, len(types), 1)

	var found bool
	for _, n := range types {
		if n.Name == "Color" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find enum Color as a type")
}

func TestCExtractor_Typedef(t *testing.T) {
	src := []byte(`typedef int MyInt;
typedef struct {
    int x;
    int y;
} Point;
`)
	e := NewCExtractor()
	result, err := e.Extract("types.c", src)
	require.NoError(t, err)

	types := nodesOfKind(result.Nodes, graph.KindType)
	require.GreaterOrEqual(t, len(types), 2)

	names := make([]string, len(types))
	for i, n := range types {
		names[i] = n.Name
	}
	assert.Contains(t, names, "MyInt")
	assert.Contains(t, names, "Point")
}

func TestCExtractor_CallSites(t *testing.T) {
	src := []byte(`#include <stdio.h>

void helper(void) {}

void greet(const char* name) {
    printf("Hello %s\n", name);
    helper();
}
`)
	e := NewCExtractor()
	result, err := e.Extract("main.c", src)
	require.NoError(t, err)

	calls := edgesOfKind(result.Edges, graph.EdgeCalls)
	require.GreaterOrEqual(t, len(calls), 2)

	var targets []string
	for _, c := range calls {
		targets = append(targets, c.To)
	}
	assert.Contains(t, targets, "unresolved::printf")
	assert.Contains(t, targets, "unresolved::helper")
}

func TestCExtractor_GlobalVariable(t *testing.T) {
	src := []byte(`int max_retries = 3;
const char* app_name = "test";

void foo(void) {
    int local = 42;
}
`)
	e := NewCExtractor()
	result, err := e.Extract("globals.c", src)
	require.NoError(t, err)

	vars := nodesOfKind(result.Nodes, graph.KindVariable)
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	assert.Contains(t, names, "max_retries")
	// local should NOT be extracted
	assert.NotContains(t, names, "local")
}

func TestCExtractor_FunctionPrototype(t *testing.T) {
	src := []byte(`int add(int a, int b);
void greet(const char* name);
`)
	e := NewCExtractor()
	result, err := e.Extract("header.h", src)
	require.NoError(t, err)

	funcs := nodesOfKind(result.Nodes, graph.KindFunction)
	require.Len(t, funcs, 2)

	names := []string{funcs[0].Name, funcs[1].Name}
	assert.Contains(t, names, "add")
	assert.Contains(t, names, "greet")
}

func TestCExtractor_FullFile(t *testing.T) {
	src := []byte(`#include <stdio.h>
#include <stdlib.h>

typedef unsigned int uint;

struct Config {
    int port;
    const char* host;
};

enum LogLevel {
    DEBUG,
    INFO,
    ERROR
};

int global_count = 0;

void process(struct Config* cfg) {
    printf("port: %d\n", cfg->port);
    global_count++;
}

int main(int argc, char* argv[]) {
    struct Config cfg;
    cfg.port = 8080;
    cfg.host = "localhost";
    process(&cfg);
    return 0;
}
`)
	e := NewCExtractor()
	result, err := e.Extract("main.c", src)
	require.NoError(t, err)

	// File node.
	files := nodesOfKind(result.Nodes, graph.KindFile)
	require.Len(t, files, 1)

	// Functions.
	funcs := nodesOfKind(result.Nodes, graph.KindFunction)
	require.Len(t, funcs, 2)

	// Types (struct Config, enum LogLevel, typedef uint).
	types := nodesOfKind(result.Nodes, graph.KindType)
	require.GreaterOrEqual(t, len(types), 3)

	// Imports.
	imports := edgesOfKind(result.Edges, graph.EdgeImports)
	assert.Len(t, imports, 2)

	// Calls from process -> printf.
	calls := edgesOfKind(result.Edges, graph.EdgeCalls)
	assert.GreaterOrEqual(t, len(calls), 1)
}
