import { Flex, IconButton, Text, useColorModeValue } from "@chakra-ui/react";
import { FiMenu } from "react-icons/fi";

export default function MobileNav({ onOpen, ...rest }) {
    return (
      <Flex
        ml={{ base: 0, md: 60 }}
        px={{ base: 4, md: 24 }}
        height="20"
        alignItems="center"
        bg={useColorModeValue("white", "black")}
        borderBottomWidth="1px"
        borderBottomColor={useColorModeValue("gray.200", "whiteAlpha.200")}
        justifyContent="flex-start"
        color={useColorModeValue("gray.900", "white")}
        position="fixed"
        top={0}
        left={0}
        right={0}
        zIndex="banner"
        {...rest}
      >
        <IconButton
          variant="outline"
          onClick={onOpen}
          aria-label="open menu"
          icon={<FiMenu />}
        />
  
        <Text fontSize="2xl" ml="8" fontWeight="bold">
          Dmytro PI
        </Text>
      </Flex>
    );
  }
