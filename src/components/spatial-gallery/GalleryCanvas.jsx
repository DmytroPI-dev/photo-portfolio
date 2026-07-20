import { Box, Stack, Text } from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { useCallback, useEffect, useRef, useState } from "react";
import { homeFloors } from "../../data/homeFloors";
import SpatialGalleryScene from "./SpatialGalleryScene";

export default function GalleryCanvas() {
  const [selectedPhotoId, setSelectedPhotoId] = useState(null);
  const [activeFloorIndex, setActiveFloorIndex] = useState(0);
  const galleryRef = useRef(null);
  const lastFloorMoveAt = useRef(0);
  const touchStartY = useRef(null);

  const activeFloor = homeFloors[activeFloorIndex] ?? homeFloors[0];

  const moveFloor = useCallback((direction) => {
    setSelectedPhotoId(null);
    setActiveFloorIndex((currentIndex) => {
      const lastFloorIndex = homeFloors.length - 1;
      return Math.min(Math.max(currentIndex + direction, 0), lastFloorIndex);
    });
  }, []);

  const jumpToFloor = useCallback((nextIndex) => {
    setSelectedPhotoId(null);
    setActiveFloorIndex(nextIndex);
  }, []);

  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        setSelectedPhotoId(null);
        return;
      }

      if (event.key === "ArrowDown" || event.key === "PageDown") {
        moveFloor(1);
      }

      if (event.key === "ArrowUp" || event.key === "PageUp") {
        moveFloor(-1);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [moveFloor]);

  const handleWheel = useCallback((event) => {
    // Trackpads emit many tiny wheel events. This throttle turns the stream into
    // one intentional floor change, which makes the gallery feel like an
    // elevator stopping at a level instead of a web page sliding by. It is bound
    // as a native non-passive listener below because React/browser passive wheel
    // handling can reject preventDefault and spam the console.
    if (Math.abs(event.deltaY) < Math.abs(event.deltaX)) return;
    event.preventDefault();

    const now = window.performance.now();
    if (now - lastFloorMoveAt.current < 720 || Math.abs(event.deltaY) < 18) {
      return;
    }

    lastFloorMoveAt.current = now;
    moveFloor(event.deltaY > 0 ? 1 : -1);
  }, [moveFloor]);

  useEffect(() => {
    const galleryElement = galleryRef.current;
    if (!galleryElement) return undefined;

    galleryElement.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      galleryElement.removeEventListener("wheel", handleWheel);
    };
  }, [handleWheel]);

  const handleTouchStart = (event) => {
    touchStartY.current = event.touches[0]?.clientY ?? null;
  };

  const handleTouchEnd = (event) => {
    if (touchStartY.current === null) return;

    const touchEndY = event.changedTouches[0]?.clientY ?? touchStartY.current;
    const deltaY = touchStartY.current - touchEndY;
    touchStartY.current = null;

    if (Math.abs(deltaY) > 42) {
      moveFloor(deltaY > 0 ? 1 : -1);
    }
  };

  return (
    <Box
      h="100%"
      minH="540px"
      bg="black"
      position="relative"
      className="home-elevator-gallery"
      ref={galleryRef}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
      sx={{ touchAction: "pan-y" }}
    >
      <Canvas
        shadows
        dpr={[1, 1.5]}
        camera={{ position: [0, 2, 15], fov: 70 }}
        gl={{ antialias: true, alpha: false }}
        onPointerMissed={() => setSelectedPhotoId(null)}
      >
        <SpatialGalleryScene
          floors={homeFloors}
          activeFloorIndex={activeFloorIndex}
          selectedPhotoId={selectedPhotoId}
          onSelect={(photoId) =>
            setSelectedPhotoId((currentPhotoId) =>
              currentPhotoId === photoId ? null : photoId
            )
          }
        />
      </Canvas>

      <Stack
        position="absolute"
        right={{ md: 5, xl: 7 }}
        top="50%"
        transform="translateY(-50%)"
        spacing={3}
        zIndex={2}
        align="center"
      >
        {homeFloors.map((floor, index) => (
          <Box
            as="button"
            key={floor.id}
            aria-label={`Go to ${floor.title}`}
            title={floor.title}
            w={index === activeFloorIndex ? "18px" : "12px"}
            h={index === activeFloorIndex ? "18px" : "12px"}
            borderRadius="999px"
            border="1px solid"
            borderColor={
              index === activeFloorIndex ? "whiteAlpha.800" : "whiteAlpha.300"
            }
            bg={floor.accent}
            opacity={index === activeFloorIndex ? 1 : 0.56}
            boxShadow={
              index === activeFloorIndex
                ? `0 0 22px ${floor.accent}`
                : "0 0 12px rgba(0, 0, 0, 0.55)"
            }
            transition="all 160ms ease"
            _hover={{
              opacity: 1,
              transform: "scale(1.15)",
            }}
            onClick={() => jumpToFloor(index)}
          />
        ))}
      </Stack>

      <Box
        position="absolute"
        left={{ md: 12, xl: 20 }}
        bottom={{ md: 9, xl: 12 }}
        maxW="360px"
        color="white"
        zIndex={2}
        pointerEvents="none"
        textShadow="0 2px 24px rgba(0, 0, 0, 0.95)"
      >
        <Text
          textTransform="uppercase"
          fontSize="xs"
          letterSpacing="0.14em"
          color="whiteAlpha.600"
        >
          Floor {activeFloorIndex + 1} / {homeFloors.length}
        </Text>
        <Text fontSize="xl" fontWeight="semibold">
          {activeFloor.title}
        </Text>
        <Text color="whiteAlpha.700" fontSize="sm">
          {activeFloor.description}
        </Text>
      </Box>
    </Box>
  );
}
