import { featuredPhotos, getPhotosByCollection } from "./photos";

const MAX_PHOTOS_PER_FLOOR = 7;

// Home floors are capped at seven pictures because the elevator camera cannot
// honestly show nine framed works without the far side images hiding behind the
// front row. Each count has its own mirrored layout so smaller collections still
// feel intentional instead of leaving one side of the room empty.
const mirroredLayouts = {
  1: [{ position: [0, 0, 1.55], rotation: [0, 0, 0], scale: 1.04 }],
  2: [
    { position: [-0.9, 0, 1.35], rotation: [0, 0.08, 0], scale: 0.96 },
    { position: [0.9, 0, 1.35], rotation: [0, -0.08, 0], scale: 0.96 },
  ],
  3: [
    { position: [0, 0, 1.55], rotation: [0, 0, 0], scale: 1 },
    { position: [-1.95, 0, 0.65], rotation: [0, Math.PI / 3.05, 0], scale: 0.9 },
    { position: [1.95, 0, 0.65], rotation: [0, -Math.PI / 3.05, 0], scale: 0.9 },
  ],
  4: [
    { position: [-0.9, 0, 1.35], rotation: [0, 0.08, 0], scale: 0.92 },
    { position: [0.9, 0, 1.35], rotation: [0, -0.08, 0], scale: 0.92 },
    { position: [-2.25, 0, 0.65], rotation: [0, Math.PI / 2.95, 0], scale: 0.86 },
    { position: [2.25, 0, 0.65], rotation: [0, -Math.PI / 2.95, 0], scale: 0.86 },
  ],
  5: [
    { position: [0, 0, 1.55], rotation: [0, 0, 0], scale: 0.96 },
    { position: [-1.35, 0, -0.45], rotation: [0, 0.03, 0], scale: 0.84 },
    { position: [1.35, 0, -0.45], rotation: [0, -0.03, 0], scale: 0.84 },
    { position: [-2.42, 0, 0.85], rotation: [0, Math.PI / 2.95, 0], scale: 0.8 },
    { position: [2.42, 0, 0.85], rotation: [0, -Math.PI / 2.95, 0], scale: 0.8 },
  ],
  6: [
    { position: [-0.88, 0, -0.55], rotation: [0, 0.03, 0], scale: 0.82 },
    { position: [0.88, 0, -0.55], rotation: [0, -0.03, 0], scale: 0.82 },
    { position: [-2.12, 0, 0.35], rotation: [0, Math.PI / 2.95, 0], scale: 0.78 },
    { position: [2.12, 0, 0.35], rotation: [0, -Math.PI / 2.95, 0], scale: 0.78 },
    { position: [-2.58, 0, 1.78], rotation: [0, Math.PI / 2.85, 0], scale: 0.72 },
    { position: [2.58, 0, 1.78], rotation: [0, -Math.PI / 2.85, 0], scale: 0.72 },
  ],
  7: [
    { position: [0, 0, 1.55], rotation: [0, 0, 0], scale: 0.96 },
    { position: [-1.18, 0, -0.55], rotation: [0, 0.03, 0], scale: 0.82 },
    { position: [1.18, 0, -0.55], rotation: [0, -0.03, 0], scale: 0.82 },
    { position: [-2.12, 0, 0.35], rotation: [0, Math.PI / 2.95, 0], scale: 0.78 },
    { position: [2.12, 0, 0.35], rotation: [0, -Math.PI / 2.95, 0], scale: 0.78 },
    { position: [-2.58, 0, 1.78], rotation: [0, Math.PI / 2.85, 0], scale: 0.72 },
    { position: [2.58, 0, 1.78], rotation: [0, -Math.PI / 2.85, 0], scale: 0.72 },
  ],
};

const createPlacements = (photos) => {
  const visiblePhotos = photos.slice(0, MAX_PHOTOS_PER_FLOOR);
  const layout = mirroredLayouts[visiblePhotos.length] ?? mirroredLayouts[7];

  return visiblePhotos.map((photo, index) => ({
    photoId: photo.id,
    position: layout[index].position,
    rotation: layout[index].rotation,
    scale: layout[index].scale,
  }));
};

const drawingPhotos = getPhotosByCollection("drawings");
const naturePhotos = getPhotosByCollection("nature");
const travelPhotos = getPhotosByCollection("travel");

export const homeFloors = [
  {
    id: "featured",
    title: "Selected Work",
    label: "Home",
    description: "A first pass through drawings, nature, and travel studies.",
    accent: "#f2b263",
    route: "/drawings",
    photos: featuredPhotos.slice(0, MAX_PHOTOS_PER_FLOOR),
  },
  {
    id: "drawings",
    title: "Drawings",
    label: "Drawings",
    description: "Hand-drawn pieces presented as the first dedicated room.",
    accent: "#daad7b",
    route: "/drawings",
    photos: drawingPhotos,
  },
  {
    id: "nature",
    title: "Nature",
    label: "Nature",
    description: "Quiet outdoor scenes and organic details.",
    accent: "#9ec6a3",
    route: "/nature",
    photos: naturePhotos,
  },
  {
    id: "travel",
    title: "Travel",
    label: "Travel",
    description: "Places, routes, and small observations from the road.",
    accent: "#9fb6df",
    route: "/travel",
    photos: travelPhotos,
  },
].map((floor, index) => ({
  ...floor,
  index,
  photos: floor.photos.slice(0, MAX_PHOTOS_PER_FLOOR),
  placements: createPlacements(floor.photos),
}));
