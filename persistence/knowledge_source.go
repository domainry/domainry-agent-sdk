package persistence

import "context"

// KnowledgeSourceRegistry is a narrow shared ownership check. Default knowledge
// must deny physical KBs claimed by a library or by private attachments, even
// after those connections are removed from host configuration. It does not
// require a caller to implement either document lifecycle repository.
type KnowledgeSourceRegistry interface {
	KnowledgeSourceManaged(context.Context, string) (bool, error)
}
