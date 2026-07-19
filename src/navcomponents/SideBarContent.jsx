import NavItem from "./NavItem";
import {
  Box,
  Button,
  Flex,
  Text,
  CloseButton,
  useColorMode,
  useColorModeValue,
} from "@chakra-ui/react";
import { MoonIcon, SunIcon } from "@chakra-ui/icons";

import { FiHome, FiTrendingUp, FiCompass, FiStar } from "react-icons/fi";

const LinkItems = [
  { name: "Home", icon: FiHome },
  { name: "Carousel", icon: FiTrendingUp },
  { name: "Testimonials", icon: FiCompass },
  { name: "Contacts", icon: FiStar },
];

export default function SideBarContent({ onClose, ...rest }) {
  const { colorMode, toggleColorMode } = useColorMode(); // Hooks for handling color mode (dark/light)
  return (
    <Box
      bg={useColorModeValue("white", "gray.900")}
      borderRight="1px"
      borderRightColor={useColorModeValue("gray.200", "gray.700")}
      w={{ base: "full", md: 60 }}
      pos="fixed"
      h="full"
      {...rest}
    >
      <Flex h="20" alignItems="center" mx="8" justifyContent="space-between">
        <Text fontSize="2xl" fontFamily="monospace" fontWeight="bold">
          Logo
        </Text>
        <CloseButton display={{ base: "flex", md: "none" }} onClick={onClose} />
      </Flex>
      {LinkItems.map((link, index) => (
        <NavItem key={link.name} icon={link.icon}>
          {link.name}
        </NavItem>
      ))}
      
      {/* Color Mode Toggle Button */}
      <Button
        fontSize="16"
        variant="ghost"
        color={useColorModeValue("gray.600", "gray.300")}
        leftIcon={colorMode === "light" ? <MoonIcon /> : <SunIcon />}
        onClick={toggleColorMode}
        justifyContent="space-between"
        alignItems="center"
        p="4"
        mx="4"
        _hover={{
          bg: "cyan.400",
          color: "white",
        }}
      >
        {colorMode === "light" ? "Dark Mode" : "Light Mode"}
      </Button>
    </Box>
  );
}
