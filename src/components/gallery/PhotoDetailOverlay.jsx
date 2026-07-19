import {
  Box,
  ButtonGroup,
  Heading,
  IconButton,
  Image,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalOverlay,
  Stack,
  Text,
  useBreakpointValue,
} from "@chakra-ui/react";
import { useEffect } from "react";
import { FiChevronLeft, FiChevronRight } from "react-icons/fi";

export default function PhotoDetailOverlay({
  isOpen,
  photo,
  onClose,
  onNext,
  onPrevious,
  showNavigation,
}) {
  const modalSize = useBreakpointValue({ base: "full", md: "6xl" });

  useEffect(() => {
    if (!isOpen || !showNavigation) return undefined;

    const handleKeyDown = (event) => {
      if (event.key === "ArrowRight") {
        event.preventDefault();
        onNext();
      }

      if (event.key === "ArrowLeft") {
        event.preventDefault();
        onPrevious();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onNext, onPrevious, showNavigation]);

  return (
    <Modal isOpen={isOpen} onClose={onClose} size={modalSize} isCentered>
      <ModalOverlay bg="blackAlpha.800" backdropFilter="blur(6px)" />
      <ModalContent
        bg="black"
        color="white"
        borderRadius={{ base: 0, md: "8px" }}
        overflow="hidden"
      >
        <ModalCloseButton zIndex="modal" />
        {photo && (
          <ModalBody p={0}>
            <Stack
              direction={{ base: "column", lg: "row" }}
              spacing={0}
              minH={{ base: "100vh", md: "70vh" }}
            >
              <Box
                flex="1"
                minH={{ base: "58vh", lg: "auto" }}
                bg="black"
                display="grid"
                placeItems="center"
              >
                <Image
                  src={photo.src}
                  alt={photo.title}
                  maxH={{ base: "58vh", lg: "82vh" }}
                  maxW="100%"
                  objectFit="contain"
                />
              </Box>

              <Stack
                w={{ base: "100%", lg: "360px" }}
                spacing={5}
                p={{ base: 5, md: 8 }}
                justify="space-between"
                bg="gray.900"
              >
                <Stack spacing={3}>
                  <Text
                    textTransform="uppercase"
                    fontSize="xs"
                    letterSpacing="0.14em"
                    color="whiteAlpha.600"
                  >
                    Selected Work
                  </Text>
                  <Heading as="h2" fontSize={{ base: "2xl", md: "3xl" }}>
                    {photo.title}
                  </Heading>
                  {photo.description && (
                    <Text color="whiteAlpha.800">{photo.description}</Text>
                  )}
                  {(photo.year || photo.location) && (
                    <Text color="whiteAlpha.600" fontSize="sm">
                      {[photo.year, photo.location].filter(Boolean).join(" / ")}
                    </Text>
                  )}
                </Stack>

                {showNavigation && (
                  <ButtonGroup spacing={3}>
                    <IconButton
                      aria-label="Previous image"
                      icon={<FiChevronLeft />}
                      onClick={onPrevious}
                      variant="outline"
                    />
                    <IconButton
                      aria-label="Next image"
                      icon={<FiChevronRight />}
                      onClick={onNext}
                      variant="outline"
                    />
                  </ButtonGroup>
                )}
              </Stack>
            </Stack>
          </ModalBody>
        )}
      </ModalContent>
    </Modal>
  );
}
