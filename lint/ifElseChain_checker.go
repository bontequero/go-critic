package lint

import (
	"go/ast"
)

func init() {
	addChecker(&ifElseChainChecker{}, attrSyntaxOnly)
}

type ifElseChainChecker struct {
	checkerBase

	cause   *ast.IfStmt
	visited map[*ast.IfStmt]bool
}

func (c *ifElseChainChecker) InitDocumentation(d *Documentation) {
	d.Summary = "Detects repeated if-else statements and suggests to replace them with switch statement"
	d.Before = `
if cond1 {
	// Code A.
} else if cond2 {
	// Code B.
} else {
	// Code C.
}`
	d.After = `
switch {
case cond1:
	// Code A.
case cond2:
	// Code B.
default:
	// Code C.
}`
	d.Note = `
Permits single else or else-if; repeated else-if or else + else-if
will trigger suggestion to use switch statement.
See docs: https://golang.org/doc/effective_go.html#switch`
}

func (c *ifElseChainChecker) EnterFunc(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	c.visited = make(map[*ast.IfStmt]bool)
	return true
}

func (c *ifElseChainChecker) VisitStmt(stmt ast.Stmt) {
	if stmt, ok := stmt.(*ast.IfStmt); ok {
		if c.visited[stmt] {
			return
		}
		c.cause = stmt
		c.checkIfStmt(stmt)
	}
}

func (c *ifElseChainChecker) checkIfStmt(stmt *ast.IfStmt) {
	const minThreshold = 2
	if c.countIfelseLen(stmt) >= minThreshold {
		c.warn()
	}
}

func (c *ifElseChainChecker) countIfelseLen(stmt *ast.IfStmt) int {
	count := 0
	for {
		switch e := stmt.Else.(type) {
		case *ast.IfStmt:
			if e.Init != nil {
				return 0 // Give up
			}
			// Else if.
			stmt = e
			count++
			c.visited[e] = true
		case *ast.BlockStmt:
			// Else branch.
			return count + 1
		default:
			// No else or else if.
			return count
		}
	}
}

func (c *ifElseChainChecker) warn() {
	c.ctx.Warn(c.cause, "should rewrite if-else to switch statement")
}
