package gallery

// NewSeedRepository mirrors the temporary frontend metadata. It gives the API
// a real, stable contract before DynamoDB exists, while the frontend keeps its
// current local-data fallback until the deployed API has been validated.
func NewSeedRepository() *MemoryRepository {
	collections := []Collection{
		{
			ID:           "drawings",
			Slug:         "drawings",
			Title:        "Drawings",
			Description:  "Hand-drawn studies and sketches. Placeholder images for now.",
			CoverPhotoID: "drawing-01",
			Order:        1,
		},
		{
			ID:           "nature",
			Slug:         "nature",
			Title:        "Nature",
			Description:  "Quiet natural scenes and organic details. Placeholder images for now.",
			CoverPhotoID: "nature-01",
			Order:        2,
		},
		{
			ID:           "travel",
			Slug:         "travel",
			Title:        "Travel",
			Description:  "Travel observations and places in motion. Placeholder images for now.",
			CoverPhotoID: "travel-01",
			Order:        3,
		},
	}

	photos := []Photo{
		seedPhoto("drawing-01", "Stillness", "Placeholder for the first hand-drawing work.", "1.jpg", "drawings", 1350, 1800, true, 1),
		seedPhoto("drawing-02", "Beautiful Beast", "Placeholder for a drawing detail or sketch.", "2.jpg", "drawings", 1800, 2400, true, 2),
		seedPhoto("drawing-03", "Morning Path", "Placeholder for a drawing series piece.", "3.jpg", "drawings", 1099, 1465, true, 3),
		seedPhoto("drawing-04", "Quiet Window", "Placeholder for a hand-drawn composition.", "4.jpg", "drawings", 1259, 1679, true, 4),
		seedPhoto("drawing-05", "Small Ceremony", "Placeholder for a drawing close-up.", "5.jpg", "drawings", 1800, 2400, true, 5),
		seedPhoto("drawing-06", "Gentle Line", "Placeholder for a final drawing in the first room.", "6.jpg", "drawings", 721, 962, true, 6),
		seedPhoto("nature-01", "Nature Study 01", "Placeholder for the nature collection.", "7.jpg", "nature", 1350, 1800, true, 7),
		seedPhoto("nature-02", "Nature Study 02", "Placeholder for natural texture or landscape work.", "8.jpg", "nature", 1800, 2400, true, 8),
		seedPhoto("nature-03", "Nature Study 03", "Placeholder for a quiet outdoor scene.", "9.jpg", "nature", 1800, 2400, true, 9),
		seedPhoto("nature-04", "Nature Study 04", "Placeholder for a nature detail.", "10.jpg", "nature", 1800, 2400, true, 10),
		seedPhoto("nature-05", "Nature Study 05", "Placeholder for an organic composition.", "11.JPG", "nature", 982, 1309, false, 11),
		seedPhoto("travel-01", "Travel Study 01", "Placeholder for the travel collection.", "12.jpg", "travel", 1519, 2026, true, 12),
		seedPhoto("travel-02", "Travel Study 02", "Placeholder for a travel observation.", "13.jpg", "travel", 1800, 2400, true, 13),
		seedPhoto("travel-03", "Travel Study 03", "Placeholder for a place or memory.", "14.JPG", "travel", 1089, 1452, false, 14),
		seedPhoto("travel-04", "Travel Study 04", "Placeholder for a travel detail.", "15.JPG", "travel", 1385, 1846, false, 15),
		seedPhoto("travel-05", "Travel Study 05", "Placeholder for a final travel image.", "16.JPG", "travel", 1267, 1690, false, 16),
	}

	return NewMemoryRepository(collections, photos)
}

func seedPhoto(id, title, description, filename, collectionID string, width, height int, featured bool, order int) Photo {
	return Photo{
		ID:           id,
		Title:        title,
		Description:  description,
		Src:          "/images/" + filename,
		CollectionID: collectionID,
		Width:        width,
		Height:       height,
		Featured:     featured,
		Order:        order,
	}
}
