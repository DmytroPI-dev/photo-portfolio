import {
  AspectRatio,
  Box,
  Heading,
  Image,
  Stack,
  Text,
} from "@chakra-ui/react";

export default function PhotoCard({ photo, onSelect }) {
  const selectPhoto = () => onSelect(photo.id);
  const handleKeyDown = (event) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      selectPhoto();
    }
  };

  return (
    <Box
      role="button"
      tabIndex={0}
      display="block"
      w="100%"
      textAlign="left"
      border="1px solid"
      borderColor="whiteAlpha.200"
      bg="whiteAlpha.100"
      borderRadius="8px"
      overflow="hidden"
      cursor="pointer"
      transition="transform 180ms ease, border-color 180ms ease, background 180ms ease"
      _hover={{
        bg: "whiteAlpha.200",
        borderColor: "whiteAlpha.400",
        transform: "translateY(-2px)",
      }}
      _focusVisible={{
        outline: "2px solid",
        outlineColor: "whiteAlpha.700",
        outlineOffset: "3px",
      }}
      onClick={selectPhoto}
      onKeyDown={handleKeyDown}
      aria-label={`Open ${photo.title}`}
    >
      <AspectRatio ratio={photo.width / photo.height}>
        <Image
          src={photo.src}
          alt={photo.title}
          objectFit="cover"
          loading="lazy"
        />
      </AspectRatio>
      <Stack spacing={1} p={{ base: 4, md: 5 }}>
        <Heading as="h2" fontSize={{ base: "md", md: "lg" }}>
          {photo.title}
        </Heading>
        {photo.description && (
          <Text color="whiteAlpha.700" fontSize="sm">
            {photo.description}
          </Text>
        )}
      </Stack>
    </Box>
  );
}
