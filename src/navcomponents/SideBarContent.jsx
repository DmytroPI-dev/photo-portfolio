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

import {
  FiAperture,
  FiEdit3,
  FiHome,
  FiMail,
  FiMap,
  FiUser,
} from "react-icons/fi";

const LinkItems = [
  { name: "Home", icon: FiHome, to: "/" },
  { name: "Drawings", icon: FiEdit3, to: "/drawings" },
  { name: "Nature", icon: FiAperture, to: "/nature" },
  { name: "Travel", icon: FiMap, to: "/travel" },
  { name: "About", icon: FiUser, to: "/about" },
  { name: "Contact", icon: FiMail, to: "/contact" },
];

export default function SideBarContent({ onClose, ...rest }) {
  const { colorMode, toggleColorMode } = useColorMode();
  return (
    <Box
      bg={useColorModeValue("white", "black")}
      borderRight="1px"
      borderRightColor={useColorModeValue("gray.200", "whiteAlpha.200")}
      w={{ base: "full", md: 60 }}
      pos="fixed"
      h="full"
      color={useColorModeValue("gray.900", "white")}
      {...rest}
    >
      <Flex h="20" alignItems="center" mx="8" justifyContent="space-between">
        <Text fontSize="2xl" fontWeight="bold">
          Dmytro PI
        </Text>
        <CloseButton display={{ base: "flex", md: "none" }} onClick={onClose} />
      </Flex>
      {LinkItems.map((link) => (
        <NavItem
          key={link.name}
          icon={link.icon}
          to={link.to}
          onClick={onClose}
        >
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
          bg: useColorModeValue("gray.100", "whiteAlpha.200"),
        }}
      >
        {colorMode === "light" ? "Dark Mode" : "Light Mode"}
      </Button>
    </Box>
  );
}
