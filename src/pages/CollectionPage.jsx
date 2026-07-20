import {
  Box,
  Heading,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import ContactSheetGallery from "../components/gallery/ContactSheetGallery";
import DrawingRoomGallery from "../components/drawing-room/DrawingRoomGallery";
import NatureGallery from "../components/nature-room/NatureGallery";
import { Navigate, useParams } from "react-router-dom";
import { getCollectionById } from "../data/collections";
import { getPhotosByCollection } from "../data/photos";

export default function CollectionPage() {
  const { collectionId } = useParams();
  const collection = getCollectionById(collectionId);

  if (!collection) {
    return <Navigate to="/" replace />;
  }

  const photos = getPhotosByCollection(collection.id);

  // Drawings and Nature now have collection-specific WebGL rooms. Travel stays
  // on the stable 2D contact sheet until it gets its own documentary/memory-wall
  // treatment.
  if (collection.id === "drawings") {
    return <DrawingRoomGallery collection={collection} photos={photos} />;
  }

  if (collection.id === "nature") {
    return <NatureGallery collection={collection} photos={photos} />;
  }

  return (
    <Box
      minH="100vh"
      bg={useColorModeValue("gray.900", "black")}
      color="white"
      px={{ base: 5, md: 10, xl: 16 }}
      py={{ base: 10, md: 16 }}
    >
      <Stack spacing={4} maxW="760px" mb={{ base: 8, md: 12 }}>
        <Text
          textTransform="uppercase"
          fontSize="sm"
          letterSpacing="0.14em"
          color="whiteAlpha.600"
        >
          Collection
        </Text>
        <Heading as="h1" fontSize={{ base: "4xl", md: "5xl" }}>
          {collection.title}
        </Heading>
        <Text color="whiteAlpha.700" fontSize={{ base: "md", md: "lg" }}>
          {collection.description}
        </Text>
      </Stack>

      <ContactSheetGallery photos={photos} />
    </Box>
  );
}
