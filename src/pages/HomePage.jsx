import {
  Box,
  Button,
  Heading,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import ContactSheetGallery from "../components/gallery/ContactSheetGallery";
import GalleryCanvas from "../components/spatial-gallery/GalleryCanvas";
import { homeFloors } from "../data/homeFloors";

export default function HomePage() {
  return (
    <Box>
      <Box
        minH={{ base: "auto", md: "100vh" }}
        bg={useColorModeValue("gray.900", "black")}
        color="white"
        position="relative"
      >
        <Box display={{ base: "none", md: "block" }} h="100vh">
          <GalleryCanvas />
        </Box>
        <Box
          display={{ base: "none", md: "block" }}
          position="absolute"
          inset={0}
          pointerEvents="none"
          bg="linear-gradient(90deg, rgba(0, 0, 0, 0.62) 0%, rgba(0, 0, 0, 0.36) 31%, rgba(0, 0, 0, 0.08) 58%, rgba(0, 0, 0, 0) 100%)"
          opacity={0.72}
        />

        <Stack
          position={{ base: "relative", md: "absolute" }}
          top={{ md: 0 }}
          left={{ md: 0 }}
          minH={{ base: "auto", md: "100vh" }}
          justify="center"
          spacing={8}
          px={{ base: 6, md: 12, xl: 20 }}
          py={{ base: 12, md: 24 }}
          maxW={{ base: "100%", md: "360px", xl: "420px" }}
          pointerEvents={{ md: "none" }}
          zIndex={1}
        >
          <Stack
            spacing={5}
            p={{ base: 0, md: 0 }}
            textShadow={{ md: "0 2px 22px rgba(0, 0, 0, 0.9)" }}
          >
            <Text
              textTransform="uppercase"
              fontSize="sm"
              letterSpacing="0.14em"
              color="whiteAlpha.700"
            >
              Dmytro PI
            </Text>
            <Heading
              as="h1"
              fontSize={{ base: "4xl", md: "3xl", xl: "4xl" }}
              lineHeight="1"
              letterSpacing="0"
            >
              A vertical gallery of drawings, nature, and travel studies.
            </Heading>
            <Text color="whiteAlpha.800" fontSize={{ base: "md", md: "md" }}>
              Hand-drawn pieces first, then quiet outdoor scenes and places in motion.
            </Text>
          </Stack>

          <Stack
            direction={{ base: "column", sm: "row" }}
            spacing={3}
            pointerEvents="auto"
          >
            <Button as={RouterLink} to="/drawings" colorScheme="whiteAlpha">
              View Drawings
            </Button>
            <Button as={RouterLink} to="/nature" variant="outline" color="white">
              Nature
            </Button>
            <Button as={RouterLink} to="/travel" variant="outline" color="white">
              Travel
            </Button>
          </Stack>
        </Stack>

        <Box
          display={{ base: "block", md: "none" }}
          px={5}
          pb={10}
          bg={useColorModeValue("gray.900", "black")}
        >
          {/* Mobile keeps the same story as the 3D Home scene, but presents each
              floor as a normal responsive image section so phones and tablets do
              not pay the cost of the desktop WebGL elevator. */}
          <Stack spacing={12}>
            {homeFloors.map((floor) => (
              <Box key={floor.id}>
                <Stack spacing={2} mb={5}>
                  <Text
                    textTransform="uppercase"
                    fontSize="xs"
                    letterSpacing="0.14em"
                    color="whiteAlpha.600"
                  >
                    {floor.label}
                  </Text>
                  <Heading as="h2" size="lg" color="white">
                    {floor.title}
                  </Heading>
                  <Text color="whiteAlpha.700">{floor.description}</Text>
                </Stack>
                <ContactSheetGallery photos={floor.photos.slice(0, 6)} />
              </Box>
            ))}
          </Stack>
        </Box>
      </Box>
    </Box>
  );
}
