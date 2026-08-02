package gallery

// Collection is the public grouping shown in the gallery navigation.
// The fields deliberately match the frontend's current local data shape so
// switching its data loader from JavaScript data to this API is mechanical.
type Collection struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CoverPhotoID string `json:"coverPhotoId"`
	Order        int    `json:"order"`
}

// Photo describes a published work. Src is a public image URL for this first
// API increment. Later it will be a deterministic CloudFront derivative URL,
// while the rest of this public contract remains stable.
type Photo struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Src          string `json:"src"`
	CollectionID string `json:"collectionId"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Year         string `json:"year"`
	Location     string `json:"location"`
	Featured     bool   `json:"featured"`
	Order        int    `json:"order"`
}

// CollectionDetail is returned by the collection route so a browser needs one
// request to render a collection page and its ordered photo set.
type CollectionDetail struct {
	Collection
	Photos []Photo `json:"photos"`
}
