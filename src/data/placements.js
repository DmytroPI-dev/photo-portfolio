import { featuredPhotos } from "./photos";

export const placements = featuredPhotos.map((photo, index) => ({
  photoId: photo.id,
  position: [index * 1.1, 0, 0],
  rotation: [0, 0, 0],
  scale: 1,
  wall: "main",
}));

export const getPlacementByPhotoId = (photoId) =>
  placements.find((placement) => placement.photoId === photoId);
