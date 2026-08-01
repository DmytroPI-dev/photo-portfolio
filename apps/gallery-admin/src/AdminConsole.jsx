import { AddIcon, ArrowBackIcon, EditIcon, ViewIcon } from "@chakra-ui/icons";
import {
  Alert,
  AlertIcon,
  Box,
  Button,
  Flex,
  Heading,
  HStack,
  IconButton,
  Link,
  Spinner,
  Stack,
  Table,
  Tbody,
  Td,
  Text,
  Textarea,
  Th,
  Thead,
  Tr,
  useToast,
  VStack,
  FormControl,
  FormLabel,
  Input,
  NumberInput,
  NumberInputField,
} from "@chakra-ui/react";
import { useCallback, useEffect, useState } from "react";
import {
  Link as RouterLink,
  Navigate,
  NavLink,
  Route,
  Routes,
  useNavigate,
  useParams,
} from "react-router-dom";
import { GalleryApiError } from "./galleryDataProvider";
import { signOut } from "./authProvider";

// The console uses only a compact set of shared primitives. It remains an
// operational workspace rather than inheriting the public portfolio's
// immersive visuals, while sharing the same Chakra component foundation.
const Page = ({ children }) => (
  <Box maxW="1240px" mx="auto" px={{ base: 5, md: 8 }} py={{ base: 7, md: 10 }}>
    {children}
  </Box>
);

const ErrorNotice = ({ error }) =>
  error ? (
    <Alert
      status="error"
      variant="left-accent"
      borderRadius="md"
      mb={6}
      alignItems="flex-start"
    >
      <AlertIcon mt={1} />
      <Text>
        {error.message || "The requested gallery data could not be loaded."}
      </Text>
    </Alert>
  ) : null;

const LoadingState = ({ label = "Loading gallery metadata..." }) => (
  <Flex
    minH="240px"
    align="center"
    justify="center"
    gap={3}
    color="whiteAlpha.700"
  >
    <Spinner color="gold.300" />
    <Text>{label}</Text>
  </Flex>
);

const ConsoleHeader = ({ user }) => {
  const navigate = useNavigate();

  const leaveConsole = async () => {
    try {
      await signOut();
    } catch {
      // The local session should be cleared even when Cognito's hosted logout
      // page cannot be reached. Returning home avoids leaving a stale console.
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

const EmptyState = ({ children }) => (
  <Box
    borderWidth="1px"
    borderStyle="dashed"
    borderColor="whiteAlpha.300"
    borderRadius="md"
    py={12}
    px={6}
    textAlign="center"
    color="whiteAlpha.700"
  >
    {children}
  </Box>
);

const CollectionForm = ({
  initialValue,
  onSubmit,
  pending,
  submitLabel,
  slugReadOnly = false,
}) => {
  const [form, setForm] = useState(initialValue);

  // Replacing a loaded record must replace form state too. This matters after a
  // version conflict refresh, and prevents an old draft from being resubmitted.
  useEffect(() => setForm(initialValue), [initialValue]);

  const change = (field) => (event) =>
    setForm((current) => ({ ...current, [field]: event.target.value }));

  return (
    <Box
      as="form"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit(form);
      }}
    >
      <Stack spacing={5} maxW="680px">
        <FormControl isRequired>
          <FormLabel>Title</FormLabel>
          <Input
            value={form.title}
            onChange={change("title")}
            autoComplete="off"
          />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Slug</FormLabel>
          <Input
            value={form.slug}
            onChange={change("slug")}
            isReadOnly={slugReadOnly}
            autoComplete="off"
          />
          {slugReadOnly ? (
            <Text mt={1} fontSize="sm" color="whiteAlpha.600">
              The URL identifier is immutable after creation.
            </Text>
          ) : null}
        </FormControl>
        <FormControl>
          <FormLabel>Description</FormLabel>
          <Textarea
            value={form.description}
            onChange={change("description")}
            rows={4}
            resize="vertical"
          />
        </FormControl>
        <FormControl>
          <FormLabel>Cover photo ID</FormLabel>
          <Input
            value={form.coverPhotoId}
            onChange={change("coverPhotoId")}
            autoComplete="off"
          />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Display order</FormLabel>
          <NumberInput
            min={1}
            value={form.order}
            onChange={(value) =>
              setForm((current) => ({ ...current, order: value }))
            }
          >
            <NumberInputField />
          </NumberInput>
        </FormControl>
        <HStack pt={2}>
          <Button
            type="submit"
            colorScheme="yellow"
            isLoading={pending}
            loadingText="Saving"
          >
            {submitLabel}
          </Button>
          <Button as={RouterLink} to="/collections" variant="solid" colorScheme="green">
            Cancel
          </Button>
        </HStack>
      </Stack>
    </Box>
  );
};

const collectionBlank = {
  title: "",
  slug: "",
  description: "",
  coverPhotoId: "",
  order: 1,
};

const CollectionsPage = ({ api }) => {
  const [collections, setCollections] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setCollections(await api.listCollections());
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <Page>
      <Flex
        justify="space-between"
        align={{ base: "flex-start", sm: "center" }}
        gap={4}
        mb={7}
        direction={{ base: "column", sm: "row" }}
      >
        <Box>
          <Heading size="lg">Collections</Heading>
          <Text mt={2} color="whiteAlpha.700">
            Organize the gallery rooms and their public metadata.
          </Text>
        </Box>
        <Button
          leftIcon={<AddIcon />}
          colorScheme="yellow"
          as={RouterLink}
          to="/collections/new"
        >
          New collection
        </Button>
      </Flex>
      <ErrorNotice error={error} />
      {loading ? <LoadingState /> : null}
      {!loading && !error && collections.length === 0 ? (
        <EmptyState>No collections have been created yet.</EmptyState>
      ) : null}
      {!loading && !error && collections.length > 0 ? (
        <Box
          overflowX="auto"
          borderWidth="1px"
          borderColor="whiteAlpha.200"
          borderRadius="md"
        >
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>Order</Th>
                <Th>Title</Th>
                <Th>Slug</Th>
                <Th>Cover photo</Th>
                <Th aria-label="Actions" />
              </Tr>
            </Thead>
            <Tbody>
              {collections.map((collection) => (
                <Tr key={collection.id} _hover={{ bg: "whiteAlpha.50" }}>
                  <Td>{collection.order}</Td>
                  <Td fontWeight="semibold">{collection.title}</Td>
                  <Td color="whiteAlpha.700">{collection.slug}</Td>
                  <Td color="whiteAlpha.700">
                    {collection.coverPhotoId || "-"}
                  </Td>
                  <Td textAlign="right">
                    <IconButton
                      aria-label={`Edit ${collection.title}`}
                      icon={<EditIcon />}
                      size="sm"
                      variant="solid"
                      colorScheme="teal"
                      onClick={() => navigate(`/collections/${collection.id}`)}
                    />
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </Box>
      ) : null}
    </Page>
  );
};

const CollectionCreatePage = ({ api }) => {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const create = async (form) => {
    setPending(true);
    setError(null);
    try {
      const collection = await api.createCollection(form);
      navigate(`/collections/${collection.id}`, { replace: true });
    } catch (reason) {
      setError(reason);
    } finally {
      setPending(false);
    }
  };

  return (
    <Page>
      <HStack spacing={3} mb={7}>
        <IconButton
          as={RouterLink}
          to="/collections"
          aria-label="Back to collections"
          icon={<ArrowBackIcon />}
          variant="solid"
          colorScheme="yellow"
        />
        <Heading size="lg">New collection</Heading>
      </HStack>
      <ErrorNotice error={error} />
      <CollectionForm
        initialValue={collectionBlank}
        onSubmit={create}
        pending={pending}
        submitLabel="Create collection"
      />
    </Page>
  );
};

const CollectionEditPage = ({ api }) => {
  const { id } = useParams();
  const [collection, setCollection] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const toast = useToast();

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setCollection(await api.getCollection(id));
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api, id]);

  useEffect(() => {
    load();
  }, [load]);

  const save = async (form) => {
    setPending(true);
    setError(null);
    try {
      const updated = await api.updateCollection(id, form);
      setCollection(updated);
      toast({
        title: "Collection saved",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
    } catch (reason) {
      if (
        reason instanceof GalleryApiError &&
        reason.code === "version_conflict"
      ) {
        toast({
          title: "Collection changed elsewhere",
          description: "The latest server version has been loaded.",
          status: "warning",
          duration: 5000,
          isClosable: true,
        });
        await load();
      } else {
        setError(reason);
      }
    } finally {
      setPending(false);
    }
  };

  if (loading)
    return (
      <Page>
        <LoadingState />
      </Page>
    );
  if (error || !collection)
    return (
      <Page>
        <ErrorNotice error={error} />
      </Page>
    );

  return (
    <Page>
      <HStack spacing={3} mb={7}>
        <IconButton
          as={RouterLink}
          to="/collections"
          aria-label="Back to collections"
          icon={<ArrowBackIcon />}
          variant="solid"
          colorScheme="yellow"
        />
        <Box>
          <Heading size="lg">Edit collection</Heading>
          <Text mt={1} color="whiteAlpha.600">
            Version {collection.version}
          </Text>
        </Box>
      </HStack>
      <ErrorNotice error={error} />
      <CollectionForm
        initialValue={collection}
        onSubmit={save}
        pending={pending}
        submitLabel="Save changes"
        slugReadOnly
      />
    </Page>
  );
};

const PhotosPage = ({ api }) => {
  const [photos, setPhotos] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    let alive = true;
    api
      .listPhotos()
      .then((items) => alive && setPhotos(items))
      .catch((reason) => alive && setError(reason))
      .finally(() => alive && setLoading(false));
    return () => {
      alive = false;
    };
  }, [api]);

  return (
    <Page>
      <Box mb={7}>
        <Heading size="lg">Photos</Heading>
        <Text mt={2} color="whiteAlpha.700">
          Inspect the current public metadata. Photo editing arrives with the
          upload pipeline.
        </Text>
      </Box>
      <ErrorNotice error={error} />
      {loading ? <LoadingState /> : null}
      {!loading && !error && photos.length === 0 ? (
        <EmptyState>No photos are available.</EmptyState>
      ) : null}
      {!loading && !error && photos.length > 0 ? (
        <Box
          overflowX="auto"
          borderWidth="1px"
          borderColor="whiteAlpha.200"
          borderRadius="md"
        >
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>Order</Th>
                <Th>Title</Th>
                <Th>Collection</Th>
                <Th>Year</Th>
                <Th>Location</Th>
                <Th>Featured</Th>
                <Th aria-label="Actions" />
              </Tr>
            </Thead>
            <Tbody>
              {photos.map((photo) => (
                <Tr key={photo.id} _hover={{ bg: "whiteAlpha.50" }}>
                  <Td>{photo.order}</Td>
                  <Td fontWeight="semibold">{photo.title}</Td>
                  <Td>{photo.collectionId}</Td>
                  <Td>{photo.year || "-"}</Td>
                  <Td>{photo.location || "-"}</Td>
                  <Td>{photo.featured ? "Yes" : "No"}</Td>
                  <Td textAlign="right">
                    <IconButton
                      aria-label={`View ${photo.title}`}
                      icon={<ViewIcon />}
                      size="sm"
                      variant="solid"
                      colorScheme="green"
                      onClick={() => navigate(`/photos/${photo.id}`)}
                    />
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </Box>
      ) : null}
    </Page>
  );
};

const PhotoDetailPage = ({ api }) => {
  const { id } = useParams();
  const [photo, setPhoto] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    api
      .getPhoto(id)
      .then((item) => alive && setPhoto(item))
      .catch((reason) => alive && setError(reason))
      .finally(() => alive && setLoading(false));
    return () => {
      alive = false;
    };
  }, [api, id]);

  if (loading)
    return (
      <Page>
        <LoadingState />
      </Page>
    );
  if (error || !photo)
    return (
      <Page>
        <ErrorNotice error={error} />
      </Page>
    );

  const fields = [
    ["Collection", photo.collectionId],
    ["Description", photo.description || "-"],
    ["Year", photo.year || "-"],
    ["Location", photo.location || "-"],
    ["Dimensions", `${photo.width} x ${photo.height}`],
    ["Featured", photo.featured ? "Yes" : "No"],
    ["Display order", photo.order],
    ["Current image URL", photo.src],
  ];
  return (
    <Page>
      <HStack spacing={3} mb={7}>
        <IconButton
          as={RouterLink}
          to="/photos"
          aria-label="Back to photos"
          icon={<ArrowBackIcon />}
          variant="solid"
          colorScheme="pink"
        />
        <Heading size="lg">{photo.title}</Heading>
      </HStack>
      <VStack
        align="stretch"
        spacing={0}
        borderWidth="1px"
        borderColor="whiteAlpha.200"
        borderRadius="md"
        overflow="hidden"
      >
        {fields.map(([label, value], index) => (
          <Box
            key={label}
            px={5}
            py={4}
            bg={index % 2 ? "whiteAlpha.50" : "transparent"}
          >
            <Text fontSize="sm" color="whiteAlpha.600">
              {label}
            </Text>
            <Text mt={1} wordBreak="break-word">
              {value}
            </Text>
          </Box>
        ))}
      </VStack>
    </Page>
  );
};

export const AdminConsole = ({ api, user }) => (
  <Box minH="100vh" bg="gray.950" color="whiteAlpha.900">
    <ConsoleHeader user={user} />
    <Routes>
      <Route path="/" element={<Navigate to="/collections" replace />} />
      <Route path="/collections" element={<CollectionsPage api={api} />} />
      <Route
        path="/collections/new"
        element={<CollectionCreatePage api={api} />}
      />
      <Route
        path="/collections/:id"
        element={<CollectionEditPage api={api} />}
      />
      <Route path="/photos" element={<PhotosPage api={api} />} />
      <Route path="/photos/:id" element={<PhotoDetailPage api={api} />} />
      <Route path="*" element={<Navigate to="/collections" replace />} />
    </Routes>
  </Box>
);
