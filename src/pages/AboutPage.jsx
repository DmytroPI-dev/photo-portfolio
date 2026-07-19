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
import { photos } from "../data/photos";

export default function AboutPage() {
  const portrait = photos[0];

  return (
    <Box
      minH="100vh"
      bg={useColorModeValue("gray.900", "black")}
      color="white"
      px={{ base: 5, md: 10, xl: 16 }}
      py={{ base: 10, md: 16 }}
    >
      <Grid
        templateColumns={{ base: "1fr", lg: "minmax(0, 0.85fr) 1fr" }}
        gap={{ base: 8, lg: 12 }}
        alignItems="center"
        maxW="1180px"
      >
        <GridItem>
          <AspectRatio ratio={4 / 5}>
            <Image
              src={portrait.src}
              alt="Artist portrait placeholder"
              objectFit="cover"
              borderRadius="8px"
            />
          </AspectRatio>
        </GridItem>
        <GridItem>
          <Stack spacing={5}>
            <Text
              textTransform="uppercase"
              fontSize="sm"
              letterSpacing="0.14em"
              color="whiteAlpha.600"
            >
              Artist
            </Text>
            <Heading as="h1" fontSize={{ base: "4xl", md: "5xl" }}>
              Dmytro PI
            </Heading>
            <Text color="whiteAlpha.800" fontSize={{ base: "md", md: "lg" }}>
              A compact introduction for the photographer and artist will live
              here: hand drawings first, then nature and travel studies. The
              portrait is a placeholder from the current image folder.
            </Text>
            <Text color="whiteAlpha.700">
              The final version can include a short artist statement, current
              location, selected tools, and the kind of work available for
              collaboration.
            </Text>
          </Stack>
        </GridItem>
      </Grid>
    </Box>
  );
}
