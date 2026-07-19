import React from "react";
import {
  Container,
  Flex,
  Box,
  Heading,
  Text,
  IconButton,
  Link,
  Button,
  VStack,
  HStack,
  Wrap,
  WrapItem,
  FormControl,
  FormLabel,
  Input,
  InputGroup,
  InputLeftElement,
  Textarea,
  useColorModeValue,
} from "@chakra-ui/react";
import { BsGithub, BsLinkedin, BsPerson } from "react-icons/bs";
import {
  MdPhone,
  MdLocationOn,
  MdFacebook,
  MdOutlineEmail,
} from "react-icons/md";

const phoneNumber = "+48733415979";
const email = "demetriy.78@gmail.com";


const ContactButton = ({ icon, label, href }) => (
  <Button
    size="lg"
    variant="ghost"
    color={useColorModeValue("gray.800", "white")}
    _hover={{ border: "2px solid #1C6FEB" }}
    leftIcon={icon}
  >
    <Link href={href} _hover={{ textDecoration: "none" }}>
      {label}
    </Link>
  </Button>
);

const SocialMediaButton = ({ label, icon }) => (
  <IconButton
    aria-label={label}
    variant="ghost"
    size="lg"
    isRound={true}
    _hover={{ bg: "#0D74FF" }}
    icon={icon}
  />
);

export default function Contacts() {
  
  return (
    <Container
      bg={useColorModeValue("gray.100", "gray.900")}
      maxW="full"
      mt={0}
      centerContent
      overflow="hidden"
    >
      <Flex>
        <Box
          bg={useColorModeValue("white", "gray.800")}
          boxShadow={"lg"}
          rounded={"xl"}
          color="white"
          borderRadius="lg"
          m={{ sm: 4, md: 16, lg: 10 }}
          p={{ sm: 5, md: 5, lg: 16 }}
        >
          <Box p={4}>
            <Wrap spacing={{ base: 20, sm: 3, md: 5, lg: 20 }}>
              <WrapItem>
                <Box>
                  <Heading color={useColorModeValue("gray.800", "white")}>
                    Contact
                  </Heading>
                  <Text
                    mt={{ sm: 3, md: 3, lg: 5 }}
                    color={useColorModeValue("gray.800", "white")}
                  >
                    Fill up the form below to contact me
                  </Text>
                  <Box py={{ base: 5, sm: 5, md: 8, lg: 10 }}>
                    <VStack pl={0} spacing={3} alignItems="flex-start">
                      <ContactButton
                        icon={<MdPhone color="#1970F1" size="20px" />}
                        label={phoneNumber}
                        href={`tel:${phoneNumber}`}
                      />
                      <ContactButton
                        icon={<MdOutlineEmail color="#1970F1" size="20px" />}
                        label={email}
                        href={`mailto:${email}`}
                      />
                      <Button
                        size="lg"
                        variant="ghost"
                        color={useColorModeValue("gray.800", "white")}
                        _hover={{ border: "2px solid #1C6FEB" }}
                        leftIcon={<MdLocationOn color="#1970F1" size="20px" />}
                      >
                        Cracow, Poland
                      </Button>
                    </VStack>
                  </Box>
                  <HStack mt={{ lg: 10, md: 10 }} spacing={5} px={5} alignItems="flex-start">
                    <SocialMediaButton
                      label="Facebook"
                      icon={<MdFacebook size="28px" />}
                    />
                    <SocialMediaButton label="GitHub" icon={<BsGithub size="28px" />} />
                    <SocialMediaButton label="LinkedIn" icon={<BsLinkedin size="28px" />} />
                  </HStack>
                </Box>
              </WrapItem>
              <WrapItem>
                <Box
                  bg={useColorModeValue("white", "gray.800")}
                  borderRadius="lg"
                >
                  <Box m={8} color={useColorModeValue("gray.800", "white")}>
                    <VStack spacing={5}>
                      <FormControl id="name">
                        <FormLabel>Your Name</FormLabel>
                        <InputGroup borderColor="#E0E1E7">
                          <InputLeftElement pointerEvents="none">
                            <BsPerson color="gray.800" />
                          </InputLeftElement>
                          <Input type="text" size="md" placeholder="Name" />
                        </InputGroup>
                      </FormControl>
                      <FormControl id="Email">
                        <FormLabel>Email</FormLabel>
                        <InputGroup borderColor="#E0E1E7">
                          <InputLeftElement pointerEvents="none">
                          <MdOutlineEmail color="gray.800" />
                          </InputLeftElement>
                          <Input type="text" size="md" placeholder="Email" />
                        </InputGroup>
                      </FormControl>
                      <FormControl id="message">
                        <FormLabel>Message</FormLabel>
                        <Textarea
                          borderColor="gray.300"
                          _hover={{
                            borderRadius: "gray.300",
                          }}
                          placeholder="Message"
                        />
                      </FormControl>
                      <FormControl id="sendMessage" float="right">
                        <Button
                          variant="solid"
                          bg="#0D74FF"
                          color={useColorModeValue("gray.800", "white")}
                          _hover={{
                            bg: "cyan.400",
                          }}
                        >
                          Send Message
                        </Button>
                      </FormControl>
                    </VStack>
                  </Box>
                </Box>
              </WrapItem>
            </Wrap>
          </Box>
        </Box>
      </Flex>
    </Container>
  );
}
