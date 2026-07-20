import { Box, Heading, Text } from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { Html, ScrollControls } from "@react-three/drei";
import { Suspense, useMemo, useState } from "react";
import DrawingRoomScene from "./DrawingRoomScene";

export default function DrawingRoomGallery({ collection, photos }) {
  const [activePhotoId, setActivePhotoId] = useState(photos[0]?.id ?? null);

  // The 3D scene owns the scroll interaction, but the page-level overlay needs
  // the currently centered photo for the title/description. The scene reports
  // that id upward through `setActivePhotoId`.
  const activePhoto = useMemo(
    () => photos.find((photo) => photo.id === activePhotoId) ?? photos[0],
    [activePhotoId, photos]
  );

  return (
    <Box
      className="drawing-room-gallery"
      position="relative"
      h={{ base: "calc(100vh - 5rem)", md: "100vh" }}
      minH={{ base: "620px", md: "720px" }}
      overflow="hidden"
      bg="#030303"
      color="white"
    >
      <Canvas
        // Shadows are intentionally disabled for this room. Circular scroll
        // renders three copies of the rail; if every duplicated spotlight casts
        // shadows, older Mac GPUs can exceed MAX_TEXTURE_IMAGE_UNITS and lose
        // the WebGL context.
        dpr={[1, 1.5]}
        camera={{ position: [0, 0.05, 6.8], fov: 52 }}
        gl={{ antialias: true, alpha: false }}
      >
        <ScrollControls
          // Drei handles the native wheel/touch scroll container. `infinite`
          // keeps the scroll position wrapping; the scene duplicates the rail
          // so the visual loop does not expose an empty gap.
          infinite
          horizontal
          pages={Math.max(photos.length + 1, 2)}
          damping={0.16}
          distance={0.78}
        >
          <Suspense
            fallback={
              <Html center style={{ color: "white", whiteSpace: "nowrap" }}>
                Loading drawings
              </Html>
            }
          >
            <DrawingRoomScene
              photos={photos}
              onActivePhotoChange={setActivePhotoId}
            />
          </Suspense>
        </ScrollControls>
      </Canvas>

      <Box
        position="absolute"
        inset={0}
        pointerEvents="none"
        // CSS vignette replaces heavier postprocessing for now. It is cheap on
        // mobile and avoids another WebGL pass while still giving the dark-room
        // edges from the reference gallery.
        bg="radial-gradient(circle at 50% 44%, transparent 0%, transparent 48%, rgba(0,0,0,0.24) 74%, rgba(0,0,0,0.58) 100%)"
      />

      <Box
        position="absolute"
        left={{ base: 5, md: 10 }}
        top={{ base: 5, md: 8 }}
        maxW={{ base: "18rem", md: "26rem" }}
        pointerEvents="none"
      >
        <Text
          textTransform="uppercase"
          letterSpacing="0.16em"
          fontSize="xs"
          color="whiteAlpha.600"
        >
          {collection.title}
        </Text>
        <Heading
          as="h1"
          mt={2}
          fontWeight="500"
          fontSize={{ base: "2xl", md: "4xl" }}
          lineHeight="1.02"
        >
          {activePhoto?.title}
        </Heading>
        {activePhoto?.description ? (
          <Text
            mt={3}
            color="whiteAlpha.700"
            fontSize={{ base: "sm", md: "md" }}
            lineHeight="1.6"
          >
            {activePhoto.description}
          </Text>
        ) : null}
      </Box>

      <Box
        position="absolute"
        left={{ base: 5, md: 10 }}
        right={{ base: 5, md: 10 }}
        bottom={{ base: 5, md: 8 }}
        h="1px"
        bg="whiteAlpha.300"
        pointerEvents="none"
      >
        <Box
          h="1px"
          // This is a simple collection progress indicator, not the physical
          // scroll percentage. With circular scroll it jumps from full back to
          // the first image, matching the active artwork id.
          w={`${photos.length ? ((photos.findIndex((photo) => photo.id === activePhotoId) + 1) / photos.length) * 100 : 0}%`}
          bg="#f2d6a2"
          transition="width 180ms ease"
        />
      </Box>
    </Box>
  );
}
