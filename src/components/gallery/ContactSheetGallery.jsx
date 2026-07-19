import { SimpleGrid } from "@chakra-ui/react";
import { useMemo, useState } from "react";
import PhotoCard from "./PhotoCard";
import PhotoDetailOverlay from "./PhotoDetailOverlay";

export default function ContactSheetGallery({ photos }) {
  const [selectedPhotoId, setSelectedPhotoId] = useState(null);

  const selectedIndex = useMemo(
    () => photos.findIndex((photo) => photo.id === selectedPhotoId),
    [photos, selectedPhotoId]
  );
  const selectedPhoto = selectedIndex >= 0 ? photos[selectedIndex] : null;

  const selectByOffset = (offset) => {
    if (!photos.length || selectedIndex < 0) return;
    const nextIndex = (selectedIndex + offset + photos.length) % photos.length;
    setSelectedPhotoId(photos[nextIndex].id);
  };

  return (
    <>
      <SimpleGrid
        columns={{ base: 1, sm: 2, xl: 3 }}
        spacing={{ base: 5, md: 6 }}
      >
        {photos.map((photo) => (
          <PhotoCard
            key={photo.id}
            photo={photo}
            onSelect={setSelectedPhotoId}
          />
        ))}
      </SimpleGrid>

      <PhotoDetailOverlay
        isOpen={Boolean(selectedPhoto)}
        photo={selectedPhoto}
        onClose={() => setSelectedPhotoId(null)}
        onNext={() => selectByOffset(1)}
        onPrevious={() => selectByOffset(-1)}
        showNavigation={photos.length > 1}
      />
    </>
  );
}
