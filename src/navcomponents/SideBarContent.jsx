import NavItem from "./NavItem";
import {
  Box,
  Flex,
  Text,
  CloseButton,
} from "@chakra-ui/react";

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
  return (
    <Box
      bg="black"
      borderRight="1px"
      borderRightColor="whiteAlpha.200"
      w={{ base: "full", md: 60 }}
      pos="fixed"
      h="full"
      color="white"
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
    </Box>
  );
}
