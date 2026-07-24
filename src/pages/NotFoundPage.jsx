import { Box, Button, Heading, Stack, Text } from "@chakra-ui/react";
import { Canvas } from "@react-three/fiber";
import { Suspense } from "react";
import { FiArrowUpRight } from "react-icons/fi";
import { Link as RouterLink } from "react-router-dom";
import NotFoundScene from "../components/not-found/NotFoundScene";

export default function NotFoundPage() {
  return (
    <Box
      position="relative"
      h={{ base: "calc(100vh - 5rem)", md: "100vh" }}
      minH={{ base: "580px", md: "680px" }}
      overflow="hidden"
      bg="#050505"
      color="white"
    >
      <Canvas
        dpr={[1, 1.5]}
        camera={{ position: [0, 0, 6.6], fov: 42 }}
        gl={{ antialias: true, alpha: false }}
      >
        <color attach="background" args={["#050505"]} />
        <fog attach="fog" args={["#050505", 5.5, 11]} />
        <Suspense fallback={null}>
          <NotFoundScene />
        </Suspense>
      </Canvas>

      <Stack
        position="absolute"
        left={{ base: 6, md: 10, xl: 14 }}
        bottom={{ base: 8, md: 12, xl: 16 }}
        maxW="24rem"
        spacing={3}
        pointerEvents="none"
        textShadow="0 2px 20px rgba(0, 0, 0, 0.9)"
      >
        <Text textTransform="uppercase" fontSize="xs" letterSpacing="0.16em" color="whiteAlpha.600">
          404 / Closed room
        </Text>
        <Heading as="h1" fontSize={{ base: "3xl", md: "4xl" }} fontWeight="500">
          This room is closed.
        </Heading>
        <Text color="whiteAlpha.700">
          The page you opened is not part of the current exhibition.
        </Text>
        <Button
          as={RouterLink}
          to="/"
          alignSelf="flex-start"
          leftIcon={<FiArrowUpRight />}
          pointerEvents="auto"
          variant="outline"
          borderColor="whiteAlpha.500"
          color="white"
          _hover={{ bg: "whiteAlpha.200" }}
        >
          Return home
        </Button>
      </Stack>
    </Box>
  );
}
