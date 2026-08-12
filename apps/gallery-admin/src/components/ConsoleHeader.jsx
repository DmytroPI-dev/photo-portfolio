import { Box, Button, Flex, Heading, HStack, Link, Text } from "@chakra-ui/react";
import { NavLink, useNavigate } from "react-router-dom";
import { signOut } from "../authProvider";

export const ConsoleHeader = ({ user }) => {
  const navigate = useNavigate();

  const leaveConsole = async () => {
    try {
      await signOut();
    } catch {
      // Clear the local session even if Cognito's hosted logout endpoint is
      // unavailable, rather than leaving a stale authenticated workspace.
      navigate("/login", { replace: true });
    }
  };

  const navItem = (to, label) => (
    <Link
      as={NavLink}
      to={to}
      px={3}
      py={2}
      borderRadius="md"
      color="whiteAlpha.700"
      _hover={{ color: "white", bg: "whiteAlpha.100" }}
      _activeLink={{ color: "black", bg: "gold.300" }}
    >
      {label}
    </Link>
  );

  return (
    <Box
      as="header"
      borderBottomWidth="1px"
      borderColor="whiteAlpha.200"
      bg="gray.950"
      position="sticky"
      top="0"
      zIndex="sticky"
    >
      <Flex
        maxW="1240px"
        mx="auto"
        minH="72px"
        px={{ base: 5, md: 8 }}
        align="center"
        gap={5}
        wrap="wrap"
      >
        <Heading size="md" letterSpacing="0">
          Gallery Administration
        </Heading>
        <HStack
          spacing={1}
          flex="1"
          order={{ base: 3, sm: 0 }}
          w={{ base: "100%", sm: "auto" }}
        >
          {navItem("/collections", "Collections")}
          {navItem("/photos", "Photos")}
        </HStack>
        <Text fontSize="sm" color="whiteAlpha.600" noOfLines={1} maxW="260px">
          {user.profile.name || user.profile.email || user.profile.username}
        </Text>
        <Button
          size="sm"
          variant="solid"
          colorScheme="telegram"
          onClick={leaveConsole}
        >
          Sign out
        </Button>
      </Flex>
    </Box>
  );
};
