import { Box, Heading, Text, Stack } from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { Html, ScrollControls } from "@react-three/drei";
import { Suspense, useMemo, useState } from "react";
import ContactSheetGallery from "../gallery/ContactSheetGallery";
import NatureScene from "./NatureScene";

export default function NatureGallery({ collection, photos }) {
  const [activePhotoId, setActivePhotoId] = useState(photos[0]?.id ?? null);

  // The 3D scene reports whichever image is nearest the center of the viewport.
  // Keeping that state here lets the HTML overlay stay crisp and accessible
  // instead of drawing readable text inside WebGL.
  const activePhoto = useMemo(
    () => photos.find((photo) => photo.id === activePhotoId) ?? photos[0],
    [activePhotoId, photos]
  );

  return (
    <>
      <Box
        display={{ base: "none", md: "block" }}
        className="nature-gallery"
        position="relative"
        h="100vh"
        minH="720px"
        overflow="hidden"
        bg="#030604"
        color="white"
      >
        <Canvas
          // Keep this lighter than the Home gallery: the scene uses only a few
          // lights and no per-image shadow maps so it should stay friendly to
          // the Mac Chrome texture-unit limit we hit earlier.
          dpr={[1, 1.5]}
          camera={{ position: [0, 0.35, 7.4], fov: 48 }}
          gl={{ antialias: true, alpha: false }}
        >
          <ScrollControls
            infinite
            horizontal
            pages={Math.max(photos.length, 2)}
            damping={0.18}
            distance={0.7}
          >
            <Suspense
              fallback={
                <Html center style={{ color: "white", whiteSpace: "nowrap" }}>
                  Loading nature
                </Html>
              }
            >
              <NatureScene
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
          bg="radial-gradient(circle at 52% 42%, transparent 0%, transparent 52%, rgba(3,8,5,0.28) 78%, rgba(0,0,0,0.62) 100%)"
        />

        <Box
          position="absolute"
          left={{ md: 10, xl: 14 }}
          top={{ md: 8, xl: 12 }}
          maxW="26rem"
          pointerEvents="none"
          textShadow="0 2px 24px rgba(0, 0, 0, 0.9)"
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
            fontSize={{ md: "4xl", xl: "5xl" }}
            lineHeight="1"
          >
            {activePhoto?.title}
          </Heading>
          {activePhoto?.description ? (
            <Text mt={3} color="whiteAlpha.700" fontSize="md" lineHeight="1.6">
              {activePhoto.description}
            </Text>
          ) : null}
        </Box>

        <Box
          position="absolute"
          left={{ md: 10, xl: 14 }}
          right={{ md: 10, xl: 14 }}
          bottom={{ md: 8, xl: 10 }}
          h="1px"
          bg="whiteAlpha.200"
          pointerEvents="none"
        >
          <Box
            h="1px"
            w={`${photos.length ? ((photos.findIndex((photo) => photo.id === activePhotoId) + 1) / photos.length) * 100 : 0}%`}
            bg="#9ec6a3"
            transition="width 180ms ease"
          />
        </Box>
      </Box>

      <Box
        display={{ base: "block", md: "none" }}
        minH="100vh"
        bg="#030604"
        color="white"
        px={5}
        py={10}
      >
        <Stack spacing={3} mb={8}>
          <Text
            textTransform="uppercase"
            fontSize="xs"
            letterSpacing="0.14em"
            color="whiteAlpha.600"
          >
            Collection
          </Text>
          <Heading as="h1" size="xl">
            {collection.title}
          </Heading>
          <Text color="whiteAlpha.700">{collection.description}</Text>
        </Stack>
        <ContactSheetGallery photos={photos} />
      </Box>
    </>
  );
}
