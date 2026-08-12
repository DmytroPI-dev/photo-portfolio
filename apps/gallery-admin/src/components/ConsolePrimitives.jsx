import {
  Alert,
  AlertIcon,
  Badge,
  Box,
  Flex,
  Spinner,
  Text,
} from "@chakra-ui/react";

// These primitives give every resource screen the same operational layout and
// recovery states without coupling the screens to one another.
export const Page = ({ children }) => (
  <Box maxW="1240px" mx="auto" px={{ base: 5, md: 8 }} py={{ base: 7, md: 10 }}>
    {children}
  </Box>
);

export const ErrorNotice = ({ error }) =>
  error ? (
    <Alert
      status="error"
      variant="left-accent"
      borderRadius="md"
      mb={6}
      alignItems="flex-start"
    >
      <AlertIcon mt={1} />
      <Text>
        {error.message || "The requested gallery data could not be loaded."}
      </Text>
    </Alert>
  ) : null;

export const LoadingState = ({ label = "Loading gallery metadata..." }) => (
  <Flex
    minH="240px"
    align="center"
    justify="center"
    gap={3}
    color="whiteAlpha.700"
  >
    <Spinner color="gold.300" />
    <Text>{label}</Text>
  </Flex>
);

export const StatusBadge = ({ status }) => {
  const colorScheme =
    {
      draft: "yellow",
      published: "green",
      archived: "gray",
    }[status] || "gray";

  return <Badge colorScheme={colorScheme}>{status}</Badge>;
};

export const EmptyState = ({ children }) => (
  <Box
    borderWidth="1px"
    borderStyle="dashed"
    borderColor="whiteAlpha.300"
    borderRadius="md"
    py={12}
    px={6}
    textAlign="center"
    color="whiteAlpha.700"
  >
    {children}
  </Box>
);
