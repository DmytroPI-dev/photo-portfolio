import { Box, Flex, Icon, useColorModeValue } from "@chakra-ui/react";
import { Link as RouterLink, useMatch } from "react-router-dom";

export default function NavItem({ icon, children, to, ...rest }) {
  const isActive = useMatch({ path: to, end: to === "/" });
  const activeBg = useColorModeValue("gray.100", "whiteAlpha.200");
  const hoverBg = useColorModeValue("gray.100", "whiteAlpha.200");

    return (
      <Box
        as={RouterLink}
        to={to}
        style={{ textDecoration: "none" }}
        _focus={{ boxShadow: "none" }}
      >
        <Flex
          align="center"
          p="4"
          mx="4"
          borderRadius="lg"
          role="group"
          cursor="pointer"
          bg={isActive ? activeBg : "transparent"}
          _hover={{
            bg: hoverBg,
          }}
          {...rest}
        >
          {icon && (
            <Icon
              mr="4"
              fontSize="16"
              _groupHover={{
                color: "white",
              }}
              as={icon}
            />
          )}
          {children}
        </Flex>
      </Box>
    );
  }
