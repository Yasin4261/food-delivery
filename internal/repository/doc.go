// Package repository holds the driven adapters: PostgreSQL implementations of
// the repository interfaces declared in the domain package.
//
// Hexagonal rule: this is the only place raw SQL lives. Each type here
// implements an interface owned by the domain package and is wired into the
// core in the composition root (cmd/api).
package repository
