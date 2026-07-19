import { Box, Button, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";

export default function NotFoundPage() {
  return (
    <Box minH="100vh" bg="black" color="white" px={6} py={20}>
      <Stack spacing={5} maxW="640px">
        <Text color="whiteAlpha.600">404</Text>
        <Heading as="h1">This room is closed.</Heading>
        <Text color="whiteAlpha.700">
          The page you opened is not part of the current exhibition.
        </Text>
        <Button as={RouterLink} to="/" alignSelf="flex-start">
          Return Home
        </Button>
      </Stack>
    </Box>
  );
}
