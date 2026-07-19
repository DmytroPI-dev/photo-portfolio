import {
  Box,
  Button,
  Heading,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import HorizontalGallery from "../components/HorizontalGallery";

export default function HomePage() {
  return (
    <Box>
      <Stack
        minH={{ base: "calc(100vh - 80px)", md: "100vh" }}
        justify="center"
        spacing={8}
        px={{ base: 6, md: 12, xl: 20 }}
        py={{ base: 16, md: 24 }}
        bg={useColorModeValue("gray.900", "black")}
        color="white"
      >
        <Stack spacing={5} maxW="760px">
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
            fontSize={{ base: "4xl", md: "6xl" }}
            lineHeight="1"
            letterSpacing="0"
          >
            Drawings, nature, and travel in a cinematic dark room.
          </Heading>
          <Text color="whiteAlpha.800" fontSize={{ base: "md", md: "lg" }}>
            A portfolio reboot built around quiet image-first rooms, starting
            with hand-drawn work and moving into outdoor and travel studies.
          </Text>
        </Stack>

        <Stack direction={{ base: "column", sm: "row" }} spacing={3}>
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

      <Box h={{ base: "70vh", md: "90vh" }}>
        <HorizontalGallery />
      </Box>
    </Box>
  );
}
