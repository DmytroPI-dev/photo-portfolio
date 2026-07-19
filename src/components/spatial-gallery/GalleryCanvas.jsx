import { Box } from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { useEffect, useMemo, useState } from "react";
import { featuredPhotos } from "../../data/photos";
import SpatialGalleryScene from "./SpatialGalleryScene";

export default function GalleryCanvas() {
  const [selectedPhotoId, setSelectedPhotoId] = useState(null);
  const photos = useMemo(() => featuredPhotos.slice(0, 9), []);

  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        setSelectedPhotoId(null);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return (
    <Box h="100%" minH="540px" bg="black">
      <Canvas
        shadows
        dpr={[1, 1.5]}
        camera={{ position: [0, 2, 15], fov: 70 }}
        gl={{ antialias: true, alpha: false }}
        onPointerMissed={() => setSelectedPhotoId(null)}
      >
        <SpatialGalleryScene
          photos={photos}
          selectedPhotoId={selectedPhotoId}
          onSelect={(photoId) =>
            setSelectedPhotoId((currentPhotoId) =>
              currentPhotoId === photoId ? null : photoId
            )
          }
        />
      </Canvas>
    </Box>
  );
}
