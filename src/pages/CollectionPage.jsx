import {
  Box,
  Heading,
  Stack,
  Text,
} from "@chakra-ui/react";
import ContactSheetGallery from "../components/gallery/ContactSheetGallery";
import DrawingRoomGallery from "../components/drawing-room/DrawingRoomGallery";
import NatureGallery from "../components/nature-room/NatureGallery";
import NotFoundPage from "./NotFoundPage";
import { useParams } from "react-router-dom";
import { useGalleryData } from "../data/galleryData";

export default function CollectionPage() {
  const { collectionId } = useParams();
  const { collections, photos, isLoading } = useGalleryData();
  const collection = collections.find(
    (candidate) => candidate.id === collectionId || candidate.slug === collectionId
  );

  if (!collection) {
    if (isLoading) {
      return <Box minH="100vh" bg="black" aria-busy="true" />;
    }

    // `/:collectionId` also matches unknown one-segment URLs. Render the 404
    // page here instead of redirecting, so the address bar keeps the URL the
    // visitor attempted to open.
    return <NotFoundPage />;
  }

  const collectionPhotos = photos
    .filter((photo) => photo.collectionId === collection.id)
    .sort((left, right) => left.order - right.order);

  // Drawings and Nature now have collection-specific WebGL rooms. Travel stays
  // on the stable 2D contact sheet until it gets its own documentary/memory-wall
  // treatment.
  if (collection.id === "drawings") {
    return <DrawingRoomGallery collection={collection} photos={collectionPhotos} />;
  }

  if (collection.id === "nature") {
    return <NatureGallery collection={collection} photos={collectionPhotos} />;
  }

  return (
    <Box
      minH="100vh"
      bg="black"
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

      <ContactSheetGallery photos={collectionPhotos} />
    </Box>
  );
}
