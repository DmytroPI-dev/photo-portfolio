import {
  AspectRatio,
  Box,
  Grid,
  GridItem,
  Heading,
  Image,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
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

      <Grid
        templateColumns={{
          base: "1fr",
          sm: "repeat(2, minmax(0, 1fr))",
          xl: "repeat(3, minmax(0, 1fr))",
        }}
        gap={{ base: 5, md: 6 }}
      >
        {photos.map((photo) => (
          <GridItem key={photo.id}>
            <Box
              border="1px solid"
              borderColor="whiteAlpha.200"
              bg="whiteAlpha.100"
              borderRadius="8px"
              overflow="hidden"
            >
              <AspectRatio ratio={photo.width / photo.height}>
                <Image src={photo.src} alt={photo.title} objectFit="cover" />
              </AspectRatio>
              <Stack spacing={1} p={4}>
                <Heading as="h2" fontSize="md">
                  {photo.title}
                </Heading>
                {photo.description && (
                  <Text color="whiteAlpha.700" fontSize="sm">
                    {photo.description}
                  </Text>
                )}
              </Stack>
            </Box>
          </GridItem>
        ))}
      </Grid>
    </Box>
  );
}
