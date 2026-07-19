import {
  Box,
  Heading,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import ContactSheetGallery from "../components/gallery/ContactSheetGallery";
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
