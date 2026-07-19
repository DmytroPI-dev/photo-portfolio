import { Box, useColorModeValue } from "@chakra-ui/react";
import Contacts from "../components/Contacts";

export default function ContactPage() {
  return (
    <Box minH="100vh" bg={useColorModeValue("gray.900", "black")}>
      <Contacts />
    </Box>
  );
}
