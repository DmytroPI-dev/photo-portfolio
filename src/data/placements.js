import { featuredPhotos } from "./photos";

const galleryPositions = [
  { position: [0, 0, 1.5], rotation: [0, 0, 0] },
  { position: [-0.8, 0, -0.6], rotation: [0, 0, 0] },
  { position: [0.8, 0, -0.6], rotation: [0, 0, 0] },
  { position: [-1.75, 0, 0.25], rotation: [0, Math.PI / 2.5, 0] },
  { position: [-2.15, 0, 1.5], rotation: [0, Math.PI / 2.5, 0] },
  { position: [-2, 0, 2.75], rotation: [0, Math.PI / 2.5, 0] },
  { position: [1.75, 0, 0.25], rotation: [0, -Math.PI / 2.5, 0] },
  { position: [2.15, 0, 1.5], rotation: [0, -Math.PI / 2.5, 0] },
  { position: [2, 0, 2.75], rotation: [0, -Math.PI / 2.5, 0] },
];

export const placements = featuredPhotos.map((photo, index) => {
  const fallbackRow = Math.floor(index / 3);
  const fallbackColumn = index % 3;
  const layout = galleryPositions[index] ?? {
    position: [fallbackColumn * 1.8 - 1.05, 0, -7 - fallbackRow * 1.8],
    rotation: [0, 0, 0],
  };

  return {
    photoId: photo.id,
    position: layout.position,
    rotation: layout.rotation,
    scale: 1,
    wall: "main",
  };
});

export const getPlacementByPhotoId = (photoId) =>
  placements.find((placement) => placement.photoId === photoId);
