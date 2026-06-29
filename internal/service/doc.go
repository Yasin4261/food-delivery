// Package service contains the application use cases. A service orchestrates a
// request by applying domain rules and calling repositories through their
// domain interfaces (ports).
//
// Hexagonal rule: no SQL and no HTTP concerns belong here. A service depends on
// domain interfaces, which makes it unit-testable with fake repositories.
package service
