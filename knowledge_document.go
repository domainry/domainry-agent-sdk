package agentsdk

import (
	"context"

	knowledgesdk "github.com/domainry/domainry-knowledge-sdk"
)

const (
	KnowledgeDocumentMaximumBytes     = knowledgesdk.KnowledgeDocumentMaximumBytes
	KnowledgeDocumentUploadChunkBytes = knowledgesdk.KnowledgeDocumentUploadChunkBytes
)

type KnowledgeDocument = knowledgesdk.KnowledgeDocument
type KnowledgeDocumentUpload = knowledgesdk.KnowledgeDocumentUpload
type KnowledgeDocumentUploadChunk = knowledgesdk.KnowledgeDocumentUploadChunk
type KnowledgeDocumentUploadProgress = knowledgesdk.KnowledgeDocumentUploadProgress
type KnowledgeDocumentSourceAccess = knowledgesdk.KnowledgeDocumentSourceAccess
type KnowledgeDocumentSourceAuthorizer = knowledgesdk.KnowledgeDocumentSourceAuthorizer
type KnowledgeDocumentSourceUploadService = knowledgesdk.KnowledgeDocumentSourceUploadService
type KnowledgeDocumentPage = knowledgesdk.KnowledgeDocumentPage
type KnowledgeDocumentDownload = knowledgesdk.KnowledgeDocumentDownload
type KnowledgeDocumentService = knowledgesdk.KnowledgeDocumentService
type KnowledgeDocumentChunkUploadService = knowledgesdk.KnowledgeDocumentChunkUploadService
type KnowledgeDocumentRetryService = knowledgesdk.KnowledgeDocumentRetryService
type KnowledgeDocumentReindexService = knowledgesdk.KnowledgeDocumentReindexService
type KnowledgeDocumentStorageScope = knowledgesdk.KnowledgeDocumentStorageScope
type KnowledgeDocumentStorage = knowledgesdk.KnowledgeDocumentStorage
type KnowledgeDocumentChunkStorage = knowledgesdk.KnowledgeDocumentChunkStorage
type KnowledgeDocumentPassage = knowledgesdk.KnowledgeDocumentPassage
type KnowledgeDocumentSearch = knowledgesdk.KnowledgeDocumentSearch
type ManagedKnowledgeDocumentSearchSource = knowledgesdk.ManagedKnowledgeDocumentSearchSource
type ManagedKnowledgeDocumentPermissionSource = knowledgesdk.ManagedKnowledgeDocumentPermissionSource
type ManagedKnowledgeDocumentSource = knowledgesdk.ManagedKnowledgeDocumentSource
type KnowledgeDocumentTransfer = knowledgesdk.KnowledgeDocumentTransfer
type KnowledgeDocumentTransferService = knowledgesdk.KnowledgeDocumentTransferService

// Import creates an independent Knowledge document from an Agent-owned
// conversation attachment. Agent owns the source attachment and authorization;
// Knowledge owns the resulting document.
type KnowledgeAttachmentImport struct {
	ClientID         string `json:"client_id"`
	ConversationID   string `json:"conversation_id"`
	AttachmentID     string `json:"attachment_id"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type KnowledgeAttachmentImportService interface {
	ImportConversationAttachment(context.Context, string, KnowledgeAttachmentImport, ConversationAuthority) (KnowledgeDocument, error)
}
