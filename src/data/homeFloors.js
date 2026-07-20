import { featuredPhotos, getPhotosByCollection } from "./photos";

// Home uses the same wall-like arrangement on every "floor", then moves the
// whole stack vertically. Keeping the layout data here makes it easy to tune a
// floor without digging through React Three Fiber scene code.
const wraparoundWallLayout = [
  { position: [0, 0, 1.55], rotation: [0, 0, 0], scale: 1.06 },
  { position: [-1.05, 0, -0.55], rotation: [0, 0, 0], scale: 0.96 },
  { position: [1.05, 0, -0.55], rotation: [0, 0, 0], scale: 0.96 },
  { position: [-2.15, 0, 0.55], rotation: [0, Math.PI / 2.65, 0], scale: 0.95 },
  { position: [-2.45, 0, 1.85], rotation: [0, Math.PI / 2.65, 0], scale: 0.9 },
  { position: [2.15, 0, 0.55], rotation: [0, -Math.PI / 2.65, 0], scale: 0.95 },
  { position: [2.45, 0, 1.85], rotation: [0, -Math.PI / 2.65, 0], scale: 0.9 },
  { position: [-0.85, 0, -2.45], rotation: [0, 0, 0], scale: 0.86 },
  { position: [0.85, 0, -2.45], rotation: [0, 0, 0], scale: 0.86 },
];

const createPlacements = (photos) =>
  photos.map((photo, index) => {
    const layout = wraparoundWallLayout[index % wraparoundWallLayout.length];
    const row = Math.floor(index / wraparoundWallLayout.length);

    return {
      photoId: photo.id,
      position: [
        layout.position[0],
        layout.position[1],
        layout.position[2] - row * 1.8,
      ],
      rotation: layout.rotation,
      scale: layout.scale,
    };
  });

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
    photos: featuredPhotos.slice(0, 9),
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
  placements: createPlacements(floor.photos),
}));
