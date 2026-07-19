export const collections = [
  {
    id: "drawings",
    slug: "drawings",
    title: "Drawings",
    description: "Hand-drawn studies and sketches. Placeholder images for now.",
    coverPhotoId: "drawing-01",
    order: 1,
  },
  {
    id: "nature",
    slug: "nature",
    title: "Nature",
    description: "Quiet natural scenes and organic details. Placeholder images for now.",
    coverPhotoId: "nature-01",
    order: 2,
  },
  {
    id: "travel",
    slug: "travel",
    title: "Travel",
    description: "Travel observations and places in motion. Placeholder images for now.",
    coverPhotoId: "travel-01",
    order: 3,
  },
];

export const getCollectionById = (collectionId) =>
  collections.find((collection) => collection.id === collectionId);
