import {
  Box,
  Heading,
  IconButton,
  Stack,
  Text,
  VisuallyHidden,
} from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { useLoader } from "@react-three/fiber";
import { Html } from "@react-three/drei";
import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { FiChevronLeft, FiChevronRight, FiMinimize2 } from "react-icons/fi";
import * as THREE from "three";
import ContactSheetGallery from "../gallery/ContactSheetGallery";
import NatureScene from "./NatureScene";

const TRANSITION_MS = 620;
const WHEEL_NAVIGATION_COOLDOWN_MS = TRANSITION_MS * 2;

const wrapIndex = (index, total) => ((index % total) + total) % total;

export default function NatureGallery({ collection, photos }) {
  const [activeIndex, setActiveIndex] = useState(0);
  const [visibleIndex, setVisibleIndex] = useState(0);
  const [direction, setDirection] = useState(1);
  const [phase, setPhase] = useState("idle");
  const [isExpanded, setIsExpanded] = useState(false);
  const transitionTimers = useRef([]);
  const lastWheelNavigationAt = useRef(0);

  const visiblePhoto = photos[visibleIndex] ?? photos[0];
  const activePhoto = photos[activeIndex] ?? photos[0];
  const canNavigate = photos.length > 1 && phase === "idle";

  const goTo = useCallback(
    (nextDirection) => {
      if (!photos.length || !canNavigate) return;

      const nextIndex = wrapIndex(activeIndex + nextDirection, photos.length);
      setDirection(nextDirection);
      setIsExpanded(false);
      setPhase("leaving");

      transitionTimers.current.forEach((timerId) => window.clearTimeout(timerId));
      transitionTimers.current = [];

      // The outgoing frame flies away first. After the old image is mostly out
      // of view, swap data and let the new frame fly in from the opposite side.
      transitionTimers.current.push(window.setTimeout(() => {
        setActiveIndex(nextIndex);
        setVisibleIndex(nextIndex);
        setPhase("entering");
      }, TRANSITION_MS));

      transitionTimers.current.push(window.setTimeout(() => {
        setPhase("idle");
        transitionTimers.current = [];
      }, TRANSITION_MS * 2));
    },
    [activeIndex, canNavigate, photos.length]
  );

  useEffect(() => {
    // Preload the foreground WebGL textures before navigation. Without this,
    // swapping to a never-seen image can briefly suspend the Canvas and look
    // like a blink before the frame flies in.
    photos.forEach((photo) => {
      useLoader.preload(THREE.TextureLoader, photo.src);
    });
  }, [photos]);

  useEffect(
    () => () => {
      transitionTimers.current.forEach((timerId) => window.clearTimeout(timerId));
    },
    []
  );

  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === "ArrowLeft") goTo(-1);
      if (event.key === "ArrowRight") goTo(1);
      if (event.key === "Escape") setIsExpanded(false);
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [goTo]);

  const handleWheel = useCallback(
    (event) => {
      if (!canNavigate || Math.abs(event.deltaY) < 18) return;

      const now = window.performance.now();
      if (now - lastWheelNavigationAt.current < WHEEL_NAVIGATION_COOLDOWN_MS) {
        return;
      }

      event.preventDefault();
      lastWheelNavigationAt.current = now;
      goTo(event.deltaY > 0 ? 1 : -1);
    },
    [canNavigate, goTo]
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
        bg="#030403"
        color="white"
        onWheel={handleWheel}
      >
        <Box
          position="absolute"
          inset={0}
          bgImage={`url("${visiblePhoto?.src}")`}
          bgSize="cover"
          bgPos="center"
          filter="blur(8px) saturate(0.78) brightness(0.54)"
          transform="scale(1.08)"
          transition={`background-image ${TRANSITION_MS}ms ease, filter ${TRANSITION_MS}ms ease`}
        />

        <Box
          position="absolute"
          inset={0}
          bg={[
            "radial-gradient(circle at 50% 44%, rgba(255,255,255,0.1), transparent 28%, rgba(0,0,0,0.62) 72%, rgba(0,0,0,0.92) 100%)",
            "linear-gradient(90deg, rgba(0,0,0,0.72), rgba(0,0,0,0.18) 35%, rgba(0,0,0,0.24) 65%, rgba(0,0,0,0.8))",
          ].join(",")}
          pointerEvents="none"
        />

        <Canvas
          // The background is HTML/CSS, so WebGL only has to draw one active
          // framed photo. That keeps Nature lighter than the earlier corridor
          // and avoids the Mac Chrome texture-unit issue from the drawings room.
          dpr={[1, 1.5]}
          camera={{ position: [0, 0.2, 7.6], fov: 42 }}
          gl={{ antialias: true, alpha: true }}
        >
          <Suspense
            fallback={
              <Html center style={{ color: "white", whiteSpace: "nowrap" }}>
                Loading nature
              </Html>
            }
          >
            <NatureScene
              photo={visiblePhoto}
              direction={direction}
              phase={phase}
              isExpanded={isExpanded}
              onToggleExpanded={() => setIsExpanded((value) => !value)}
            />
          </Suspense>
        </Canvas>

        <Stack
          position="absolute"
          left={{ md: 10, xl: 14 }}
          top={{ md: 8, xl: 12 }}
          maxW="27rem"
          spacing={3}
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
            fontWeight="500"
            fontSize={{ md: "4xl", xl: "5xl" }}
            lineHeight="1"
          >
            {activePhoto?.title}
          </Heading>
          <Text color="whiteAlpha.700" fontSize="md" lineHeight="1.6">
            {activePhoto?.description}
          </Text>
        </Stack>

        <IconButton
          aria-label="Previous nature photo"
          icon={<FiChevronLeft />}
          position="absolute"
          left={{ md: 6, xl: 8 }}
          top="50%"
          transform="translateY(-50%)"
          variant="ghost"
          color="white"
          colorScheme="whiteAlpha"
          fontSize="2xl"
          isDisabled={!canNavigate}
          onClick={() => goTo(-1)}
        />

        <IconButton
          aria-label="Next nature photo"
          icon={<FiChevronRight />}
          position="absolute"
          right={{ md: 6, xl: 8 }}
          top="50%"
          transform="translateY(-50%)"
          variant="ghost"
          color="white"
          colorScheme="whiteAlpha"
          fontSize="2xl"
          isDisabled={!canNavigate}
          onClick={() => goTo(1)}
        />

        {isExpanded ? (
          <IconButton
            aria-label="Close enlarged photo"
            icon={<FiMinimize2 />}
            position="absolute"
            right={{ md: 8, xl: 10 }}
            bottom={{ md: 14, xl: 16 }}
            variant="outline"
            color="white"
            borderColor="whiteAlpha.400"
            bg="blackAlpha.300"
            _hover={{ bg: "whiteAlpha.200" }}
            onClick={() => setIsExpanded(false)}
          />
        ) : null}

        <VisuallyHidden>
          Use left and right arrow keys to move through Nature images. Click the
          framed photo to enlarge or collapse it.
        </VisuallyHidden>
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
