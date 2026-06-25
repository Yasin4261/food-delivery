// Package domain is the application core. It holds business entities and the
// repository interfaces (ports) the core depends on.
//
// Hexagonal rule: this package must not import any other internal package, any
// database driver, or any web framework. Dependencies point inward, so the
// outer layers (service, repository, handler) import domain, never the reverse.
package domain
