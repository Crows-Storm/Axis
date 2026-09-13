package event

import "context"

// The MongoRepository Mongo Database from service, Document is: domain_event
// e.g.
// User service -> database(axis-user) -> document(domain_event)
// Auth service -> database(axis-auth) -> document(domain_event)
//
// store domain all the Snapshot
// When searching for a specific domain, all snapshot records will be retrieved,
// and event backtracking will be performed based on the domain root version.
type MongoRepository interface {
	Save(ctx context.Context, event DomainEvent) error
	BuildDomain(ctx context.Context, event DomainEvent) error
}
