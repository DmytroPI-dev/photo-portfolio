package gallery

// PublicationStatus controls whether a canonical record is represented in the
// anonymous public read model. Draft and archived records remain visible only
// through future Cognito-protected admin routes.
type PublicationStatus string

const (
	PublicationDraft     PublicationStatus = "draft"
	PublicationPublished PublicationStatus = "published"
	PublicationArchived  PublicationStatus = "archived"
)

// ProcessingStatus is independent from publication. A finished image can be
// held as a draft, and a published photo keeps serving its current derivative
// while a replacement file is processing.
type ProcessingStatus string

const (
	ProcessingNotRequired ProcessingStatus = "not_required"
	ProcessingPending     ProcessingStatus = "pending"
	ProcessingProcessing  ProcessingStatus = "processing"
	ProcessingReady       ProcessingStatus = "ready"
	ProcessingFailed      ProcessingStatus = "failed"
)

// AdminCollection is the canonical, administrator-owned collection document.
// Collection remains embedded so fields shared with the public response cannot
// quietly drift into a second hand-maintained representation.
type AdminCollection struct {
	Collection
	Status    PublicationStatus `json:"status"`
	Version   int               `json:"version"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

// AdminPhoto is the canonical, administrator-owned photo document. The
// embedded public Photo is copied to public read records only when this record
// is published. Upload-specific fields deliberately stay out of public JSON.
type AdminPhoto struct {
	Photo
	Status           PublicationStatus `json:"status"`
	ProcessingStatus ProcessingStatus  `json:"processingStatus"`
	AltText          string            `json:"altText"`
	Tags             []string          `json:"tags"`
	FocalPointX      float64           `json:"focalPointX"`
	FocalPointY      float64           `json:"focalPointY"`
	OriginalKey      string            `json:"originalKey"`
	DerivativeKey    string            `json:"derivativeKey"`
	ProcessingError  string            `json:"processingError"`
	Version          int               `json:"version"`
	CreatedAt        string            `json:"createdAt"`
	UpdatedAt        string            `json:"updatedAt"`
}

const seedMetadataTimestamp = "2026-01-01T00:00:00Z"

// CanonicalSeedData gives the initial placeholder portfolio the same private
// shape that administrator-authored work will use. The timestamp is fixed so
// rerunning the bootstrap command remains deterministic and idempotent.
func CanonicalSeedData() ([]AdminCollection, []AdminPhoto) {
	collections, photos := SeedData()

	adminCollections := make([]AdminCollection, 0, len(collections))
	for _, collection := range collections {
		adminCollections = append(adminCollections, AdminCollection{
			Collection: collection,
			Status:     PublicationPublished,
			Version:    1,
			CreatedAt:  seedMetadataTimestamp,
			UpdatedAt:  seedMetadataTimestamp,
		})
	}

	adminPhotos := make([]AdminPhoto, 0, len(photos))
	for _, photo := range photos {
		adminPhotos = append(adminPhotos, AdminPhoto{
			Photo:            photo,
			Status:           PublicationPublished,
			ProcessingStatus: ProcessingNotRequired,
			AltText:          photo.Title,
			Tags:             []string{},
			Version:          1,
			CreatedAt:        seedMetadataTimestamp,
			UpdatedAt:        seedMetadataTimestamp,
		})
	}

	return adminCollections, adminPhotos
}
