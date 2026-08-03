import {
  AddIcon,
  ArrowBackIcon,
  ArrowDownIcon,
  ArrowUpIcon,
  DragHandleIcon,
  EditIcon,
  ViewIcon,
} from "@chakra-ui/icons";
import {
  Alert,
  AlertIcon,
  Badge,
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
  Image,
  Select,
  Switch,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  useDisclosure,
} from "@chakra-ui/react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  Link as RouterLink,
  Navigate,
  NavLink,
  Route,
  Routes,
  useNavigate,
  useParams,
} from "react-router-dom";
import { GalleryApiError, photoPreviewURL } from "./galleryDataProvider";
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

const CollectionStatus = ({ status }) => {
  const colorScheme =
    {
      draft: "yellow",
      published: "green",
      archived: "gray",
    }[status] || "gray";

  return <Badge colorScheme={colorScheme}>{status}</Badge>;
};

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
  readOnly = false,
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
        if (!readOnly) {
          onSubmit(form);
        }
      }}
    >
      <Stack spacing={5} maxW="680px">
        <FormControl isRequired>
          <FormLabel>Title</FormLabel>
          <Input
            value={form.title}
            onChange={change("title")}
            isReadOnly={readOnly}
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
            isReadOnly={readOnly}
            rows={4}
            resize="vertical"
          />
        </FormControl>
        <FormControl>
          <FormLabel>Cover photo ID</FormLabel>
          <Input
            value={form.coverPhotoId}
            onChange={change("coverPhotoId")}
            isReadOnly={readOnly}
            autoComplete="off"
          />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Display order</FormLabel>
          <NumberInput
            min={1}
            value={form.order}
            isReadOnly={readOnly}
            onChange={(value) =>
              setForm((current) => ({ ...current, order: value }))
            }
          >
            <NumberInputField />
          </NumberInput>
        </FormControl>
        {!readOnly ? (
          <HStack pt={2}>
            <Button
              type="submit"
              colorScheme="yellow"
              isLoading={pending}
              loadingText="Saving"
            >
              {submitLabel}
            </Button>
            <Button
              as={RouterLink}
              to="/collections"
              variant="solid"
              colorScheme="green"
            >
              Cancel
            </Button>
          </HStack>
        ) : null}
      </Stack>
    </Box>
  );
};

const DeleteCollectionControl = ({ collection, pending, onDelete }) => {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [confirmation, setConfirmation] = useState("");
  const confirmed = confirmation === collection.slug;

  const close = () => {
    setConfirmation("");
    onClose();
  };

  const remove = async () => {
    const deleted = await onDelete(confirmation);
    if (deleted) {
      close();
    }
  };

  return (
    <>
      <Button colorScheme="red" onClick={onOpen} isLoading={pending}>
        Delete
      </Button>
      <Modal isOpen={isOpen} onClose={close} isCentered>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader color="red.500">Delete collection</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Text color="black">
              This permanently removes the archived collection metadata. Type
              <Text as="span" fontWeight="bold">
                {" "}
                {collection.slug}
              </Text>{" "}
              to confirm.
            </Text>
            <Input
              mt={4}
              value={confirmation}
              color={confirmed ? "green.500" : "red.500"}
              onChange={(event) => setConfirmation(event.target.value)}
              autoComplete="off"
              aria-label="Collection slug confirmation"
            />
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={close}>
              Cancel
            </Button>
            <Button
              colorScheme="red"
              onClick={remove}
              isDisabled={!confirmed || pending}
              isLoading={pending}
            >
              Permanently delete
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
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
                <Th>Status</Th>
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
                  <Td>
                    <CollectionStatus status={collection.status} />
                  </Td>
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
  const navigate = useNavigate();

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
      await handleMutationError(reason);
    } finally {
      setPending(false);
    }
  };

  const handleMutationError = async (reason) => {
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
  };

  const transition = async (action) => {
    setPending(true);
    setError(null);
    try {
      const updated =
        action === "publish"
          ? await api.publishCollection(id, collection.version)
          : action === "archive"
            ? await api.archiveCollection(id, collection.version)
            : await api.restoreCollection(id, collection.version);
      setCollection(updated);
      toast({
        title:
          action === "publish"
            ? "Collection published"
            : action === "archive"
              ? "Collection archived"
              : "Collection restored as draft",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
    } catch (reason) {
      await handleMutationError(reason);
    } finally {
      setPending(false);
    }
  };

  const remove = async (confirmationSlug) => {
    setPending(true);
    setError(null);
    try {
      await api.deleteCollection(id, collection.version, confirmationSlug);
      toast({
        title: "Collection permanently deleted",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
      navigate("/collections", { replace: true });
      return true;
    } catch (reason) {
      await handleMutationError(reason);
      return false;
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
  // A failed save or lifecycle action should leave the loaded collection on
  // screen. The alert appears above the form, while the back control and the
  // current metadata stay available for the administrator to recover.
  if (!collection)
    return (
      <Page>
        <ErrorNotice error={error} />
      </Page>
    );

  return (
    <Page>
      <Flex
        justify="space-between"
        align={{ base: "flex-start", sm: "center" }}
        gap={4}
        mb={7}
        direction={{ base: "column", sm: "row" }}
      >
        <HStack spacing={3}>
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
            <HStack mt={1} spacing={3}>
              <CollectionStatus status={collection.status} />
              <Text color="whiteAlpha.600">Version {collection.version}</Text>
            </HStack>
          </Box>
        </HStack>
        <HStack>
          {collection.status === "draft" ? (
            <Button
              colorScheme="green"
              onClick={() => transition("publish")}
              isLoading={pending}
            >
              Publish
            </Button>
          ) : null}
          {collection.status === "draft" ||
          collection.status === "published" ? (
            <Button
              colorScheme="orange"
              onClick={() => transition("archive")}
              isLoading={pending}
            >
              Archive
            </Button>
          ) : null}
          {collection.status === "archived" ? (
            <>
              <Button
                colorScheme="blue"
                onClick={() => transition("restore")}
                isLoading={pending}
              >
                Restore
              </Button>
              <DeleteCollectionControl
                collection={collection}
                pending={pending}
                onDelete={remove}
              />
            </>
          ) : null}
        </HStack>
      </Flex>
      <ErrorNotice error={error} />
      <CollectionForm
        initialValue={collection}
        onSubmit={save}
        pending={pending}
        submitLabel="Save changes"
        slugReadOnly
        readOnly={collection.status === "archived"}
      />
    </Page>
  );
};

const PhotoStatus = ({ status }) => <CollectionStatus status={status} />;

const ProcessingStatus = ({ status }) =>
  status && status !== "not_required" ? (
    <Badge colorScheme={status === "ready" ? "green" : "yellow"}>
      {status}
    </Badge>
  ) : null;

const PhotoPreview = ({ photo, api, size = "72px" }) => {
  const [privatePreviewURL, setPrivatePreviewURL] = useState("");
  const publicPreviewURL = photo.previewURL || photoPreviewURL(photo.src);

  useEffect(() => {
    let active = true;
    if (publicPreviewURL || !api || !photo.id || !photo.originalKey) {
      setPrivatePreviewURL("");
      return () => {
        active = false;
      };
    }
    api
      .getPhotoPreview(photo.id)
      .then((response) => active && setPrivatePreviewURL(response.url))
      .catch(() => active && setPrivatePreviewURL(""));
    return () => {
      active = false;
    };
  }, [api, photo.id, photo.originalKey, publicPreviewURL]);

  const src = publicPreviewURL || privatePreviewURL;
  return src ? (
    <Image
      src={src}
      alt={photo.altText || photo.title || "Photo preview"}
      boxSize={size}
      objectFit="cover"
      borderRadius="sm"
      bg="whiteAlpha.100"
    />
  ) : (
    <Box boxSize={size} borderRadius="sm" bg="whiteAlpha.100" />
  );
};

const imageDetails = (file) =>
  new Promise((resolve, reject) => {
    const objectURL = URL.createObjectURL(file);
    const image = new window.Image();
    image.onload = () => {
      URL.revokeObjectURL(objectURL);
      resolve({ width: image.naturalWidth, height: image.naturalHeight });
    };
    image.onerror = () => {
      URL.revokeObjectURL(objectURL);
      reject(new Error("The selected file is not a readable image."));
    };
    image.src = objectURL;
  });

const titleFromFilename = (filename) =>
  filename
    .replace(/\.[^.]+$/, "")
    .replace(/[-_]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();

const PhotoDropzone = ({ onSelect, pending, readOnly = false }) => {
  const inputRef = useRef(null);
  const [dragging, setDragging] = useState(false);
  const selectFiles = (files) => {
    const imageFiles = Array.from(files || []);
    if (imageFiles.length > 0) onSelect(imageFiles);
  };

  return (
    <FormControl isRequired>
      <FormLabel>Image</FormLabel>
      <Box
        borderWidth="1px"
        borderStyle="dashed"
        borderColor={dragging ? "gold.300" : "whiteAlpha.400"}
        borderRadius="md"
        px={6}
        py={10}
        textAlign="center"
        bg={dragging ? "whiteAlpha.100" : "transparent"}
        cursor={readOnly || pending ? "default" : "pointer"}
        onClick={() => !readOnly && !pending && inputRef.current?.click()}
        onDragEnter={(event) => {
          event.preventDefault();
          if (!readOnly && !pending) setDragging(true);
        }}
        onDragOver={(event) => event.preventDefault()}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => {
          event.preventDefault();
          setDragging(false);
          if (!readOnly && !pending) selectFiles(event.dataTransfer.files);
        }}
      >
        <Stack spacing={3} align="center">
          <Text color="whiteAlpha.800">
            {pending ? "Uploading images..." : "Drop images here"}
          </Text>
          <Button
            type="button"
            size="sm"
            variant="outline"
            colorScheme="yellow"
            isDisabled={readOnly || pending}
            onClick={(event) => {
              event.stopPropagation();
              inputRef.current?.click();
            }}
          >
            Choose image
          </Button>
        </Stack>
        <Input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          display="none"
          onChange={(event) => selectFiles(event.target.files)}
        />
      </Box>
    </FormControl>
  );
};

const PhotoForm = ({
  initialValue,
  collections,
  onSubmit,
  pending,
  submitLabel,
  readOnly = false,
  api,
}) => {
  const [form, setForm] = useState(initialValue);
  const collectionOptions = collections.filter(
    (collection) =>
      collection.status !== "archived" || collection.id === form.collectionId,
  );

  useEffect(() => setForm(initialValue), [initialValue]);

  const change = (field) => (event) =>
    setForm((current) => ({ ...current, [field]: event.target.value }));
  return (
    <Box
      as="form"
      onSubmit={(event) => {
        event.preventDefault();
        if (!readOnly) {
          onSubmit({ ...form, tags: form.tagsText.split(",") });
        }
      }}
    >
      <Stack spacing={5} maxW="780px">
        {form.previewURL || form.src || form.originalKey ? (
          <PhotoPreview photo={form} api={api} size="180px" />
        ) : null}
        <FormControl isRequired>
          <FormLabel>Title</FormLabel>
          <Input
            value={form.title}
            onChange={change("title")}
            isReadOnly={readOnly}
            autoComplete="off"
          />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Collection</FormLabel>
          <Select
            value={form.collectionId}
            onChange={change("collectionId")}
            isDisabled={readOnly}
          >
            <option
              value=""
              disabled
              style={{ color: "gray" }}
            >
              Choose a collection
            </option>
            {collectionOptions.map((collection) => (
              <option
                key={collection.id}
                value={collection.id}
                disabled={collection.status === "archived"}
                style={{
                  color: collection.status === "archived" ? "gray" : "whiteAlpha.900",
                  backgroundColor: collection.status === "archived" ? "inherit" : "gray",
                }}
              >
                {collection.title} ({collection.status})
              </option>
            ))}
          </Select>
        </FormControl>
        <FormControl>
          <FormLabel>Description</FormLabel>
          <Textarea
            value={form.description}
            onChange={change("description")}
            isReadOnly={readOnly}
            rows={4}
            resize="vertical"
          />
        </FormControl>
        <FormControl>
          <FormLabel>Alt text</FormLabel>
          <Input
            value={form.altText}
            onChange={change("altText")}
            isReadOnly={readOnly}
            autoComplete="off"
          />
        </FormControl>
        <FormControl>
          <FormLabel>Tags</FormLabel>
          <Input
            value={form.tagsText}
            onChange={change("tagsText")}
            isReadOnly={readOnly}
            autoComplete="off"
          />
        </FormControl>
        <HStack align="start" spacing={4} flexWrap="wrap">
          <FormControl display="flex" alignItems="center" minW="160px" pt={8}>
            <FormLabel mb="0">Featured</FormLabel>
            <Switch
              isChecked={form.featured}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  featured: event.target.checked,
                }))
              }
              isDisabled={readOnly}
            />
          </FormControl>
        </HStack>
        <HStack align="start" spacing={4} flexWrap="wrap">
          <FormControl minW="160px" flex="1">
            <FormLabel>Year</FormLabel>
            <Input
              value={form.year}
              onChange={change("year")}
              isReadOnly={readOnly}
            />
          </FormControl>
          <FormControl minW="160px" flex="1">
            <FormLabel>Location</FormLabel>
            <Input
              value={form.location}
              onChange={change("location")}
              isReadOnly={readOnly}
            />
          </FormControl>
        </HStack>
        {!readOnly ? (
          <HStack pt={2}>
            <Button
              type="submit"
              colorScheme="yellow"
              isLoading={pending}
              loadingText="Saving"
            >
              {submitLabel}
            </Button>
            <Button
              as={RouterLink}
              to="/photos"
              variant="solid"
              colorScheme="green"
            >
              Cancel
            </Button>
          </HStack>
        ) : null}
      </Stack>
    </Box>
  );
};

const groupPhotosByCollection = (photos) =>
  photos.reduce((groups, photo) => {
    const group = groups.get(photo.collectionId) || [];
    group.push(photo);
    groups.set(photo.collectionId, group);
    return groups;
  }, new Map());

const PhotosPage = ({ api }) => {
  const [photos, setPhotos] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [pendingCollection, setPendingCollection] = useState(null);
  const [draggedID, setDraggedID] = useState(null);
  const navigate = useNavigate();

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setPhotos(await api.listPhotos());
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    load();
  }, [load]);

  const reorder = async (collectionId, collectionPhotos, fromID, toID) => {
    const fromIndex = collectionPhotos.findIndex(
      (photo) => photo.id === fromID,
    );
    const toIndex = collectionPhotos.findIndex((photo) => photo.id === toID);
    if (fromIndex < 0 || toIndex < 0 || fromIndex === toIndex) return;
    const nextOrder = [...collectionPhotos];
    nextOrder.splice(toIndex, 0, nextOrder.splice(fromIndex, 1)[0]);
    setPendingCollection(collectionId);
    setError(null);
    try {
      const updated = await api.reorderPhotos(collectionId, nextOrder);
      const byID = new Map(updated.map((photo) => [photo.id, photo]));
      setPhotos((current) =>
        current.map((photo) => byID.get(photo.id) || photo),
      );
    } catch (reason) {
      setError(reason);
    } finally {
      setPendingCollection(null);
      setDraggedID(null);
    }
  };

  const grouped = groupPhotosByCollection(photos);
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
          <Heading size="lg">Photos</Heading>
          <Text mt={2} color="whiteAlpha.700">
            Arrange works within each collection and manage publication
            metadata.
          </Text>
        </Box>
        <Button
          leftIcon={<AddIcon />}
          colorScheme="yellow"
          as={RouterLink}
          to="/photos/new"
        >
          New photo
        </Button>
      </Flex>
      <ErrorNotice error={error} />
      {loading ? <LoadingState /> : null}
      {!loading && !error && photos.length === 0 ? (
        <EmptyState>No photos have been created yet.</EmptyState>
      ) : null}
      {!loading && !error ? (
        <Stack spacing={8}>
          {[...grouped.entries()].map(([collectionId, collectionPhotos]) => (
            <Box
              key={collectionId}
              borderWidth="1px"
              borderColor="whiteAlpha.200"
              borderRadius="md"
              overflowX="auto"
            >
              <Flex
                px={5}
                py={4}
                align="center"
                justify="space-between"
                borderBottomWidth="1px"
                borderColor="whiteAlpha.200"
              >
                <Heading size="sm">{collectionId}</Heading>
                {pendingCollection === collectionId ? (
                  <Spinner size="sm" color="gold.300" />
                ) : null}
              </Flex>
              <Table variant="striped" colorScheme="whiteAlpha">
                <Thead>
                  <Tr
                    bg="gray.300"
                 >
                    <Th aria-label="Reorder" />
                    <Th>Preview</Th>
                    <Th>Order</Th>
                    <Th>Title</Th>
                    <Th>Status</Th>
                    <Th>Featured</Th>
                    <Th aria-label="Actions" />
                  </Tr>
                </Thead>
                <Tbody>
                  {collectionPhotos
                    .sort((left, right) => left.order - right.order)
                    .map((photo, index) => (
                      <Tr
                        key={photo.id}
                        draggable={!pendingCollection}
                        onDragStart={() => setDraggedID(photo.id)}
                        onDragOver={(event) => event.preventDefault()}
                        onDrop={() =>
                          reorder(
                            collectionId,
                            collectionPhotos,
                            draggedID,
                            photo.id,
                          )
                        }
                        _hover={{ bg: "whiteAlpha.50" }}
                      >
                        <Td>
                          <IconButton
                            aria-label={`Drag ${photo.title}`}
                            icon={<DragHandleIcon />}
                            size="sm"
                            variant="solid"
                            colorScheme="green.500"
                            cursor="grab"
                          />
                        </Td>
                        <Td>
                          <PhotoPreview photo={photo} api={api} size="48px" />
                        </Td>
                        <Td>{photo.order}</Td>
                        <Td fontWeight="semibold">{photo.title}</Td>
                        <Td>
                            <PhotoStatus status={photo.status} />
                            <ProcessingStatus status={photo.processingStatus} />
                        </Td>
                        <Td>{photo.featured ? "Yes" : "No"}</Td>
                        <Td textAlign="right">
                          <HStack justify="flex-end">
                            <IconButton
                              aria-label={`Move ${photo.title} up`}
                              icon={<ArrowUpIcon />}
                              size="sm"
                              variant="solid"
                              colorScheme="green"
                              isDisabled={
                                index === 0 ||
                                pendingCollection === collectionId
                              }
                              onClick={() =>
                                reorder(
                                  collectionId,
                                  collectionPhotos,
                                  photo.id,
                                  collectionPhotos[index - 1].id,
                                )
                              }
                            />
                            <IconButton
                              aria-label={`Move ${photo.title} down`}
                              icon={<ArrowDownIcon />}
                              size="sm"
                              variant="solid"
                              colorScheme="yellow"
                              isDisabled={
                                index === collectionPhotos.length - 1 ||
                                pendingCollection === collectionId
                              }
                              onClick={() =>
                                reorder(
                                  collectionId,
                                  collectionPhotos,
                                  photo.id,
                                  collectionPhotos[index + 1].id,
                                )
                              }
                            />
                            <IconButton
                              aria-label={`Edit ${photo.title}`}
                              icon={<EditIcon />}
                              size="sm"
                              variant="solid"
                              colorScheme="teal"
                              onClick={() => navigate(`/photos/${photo.id}`)}
                            />
                          </HStack>
                        </Td>
                      </Tr>
                    ))}
                </Tbody>
              </Table>
            </Box>
          ))}
        </Stack>
      ) : null}
    </Page>
  );
};

const usePhotoCollections = (api) => {
  const [collections, setCollections] = useState([]);
  const [error, setError] = useState(null);
  useEffect(() => {
    let active = true;
    api
      .listCollections()
      .then((items) => active && setCollections(items))
      .catch((reason) => active && setError(reason));
    return () => {
      active = false;
    };
  }, [api]);
  return { collections, collectionError: error };
};

const PhotoCreatePage = ({ api }) => {
  const { collections, collectionError } = usePhotoCollections(api);
  const [collectionId, setCollectionId] = useState("");
  const [pending, setPending] = useState(false);
  const [uploadedPhotos, setUploadedPhotos] = useState([]);
  const [error, setError] = useState(null);

  const upload = async (files) => {
    if (!collectionId) {
      setError(new Error("Choose a collection before adding images."));
      return;
    }
    setPending(true);
    setError(null);
    try {
      // Upload sequentially so each small collection stays comfortably inside
      // the API and DynamoDB ordering limits, while still accepting a batch.
      for (const file of files) {
        const [details, uploadDetails] = await Promise.all([
          imageDetails(file),
          api.uploadOriginal(file),
        ]);
        const title = titleFromFilename(file.name) || "Untitled photo";
        const photo = await api.createPhoto({
          uploadId: uploadDetails.photoId,
          originalKey: uploadDetails.originalKey,
          title,
          description: "",
          collectionId,
          width: details.width,
          height: details.height,
          year: "",
          location: "",
          featured: false,
          order: 0,
          altText: title,
          tags: [],
        });
        // Preserve each completed draft immediately. If a later file fails,
        // the operator can see the successful uploads instead of retrying the
        // entire batch and accidentally creating duplicates.
        setUploadedPhotos((current) => [photo, ...current]);
      }
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
          to="/photos"
          aria-label="Back to photos"
          icon={<ArrowBackIcon />}
          variant="solid"
          colorScheme="yellow"
        />
        <Heading size="lg">Add photos</Heading>
      </HStack>
      <ErrorNotice error={error || collectionError} />
      <Stack spacing={5} maxW="780px">
        <FormControl isRequired>
          <FormLabel>Collection</FormLabel>
          <Select
            value={collectionId}
            onChange={(event) => setCollectionId(event.target.value)}
            isDisabled={pending}
          >
            <option value="" disabled style={{ color: "gray" }}>
              Choose a collection
            </option>
            {collections
              .filter((collection) => collection.status !== "archived")
              .map((collection) => (
                <option
                  key={collection.id}
                  value={collection.id}
                  style={{ color: "white", backgroundColor: "#1f2937" }}
                >
                  {collection.title}
                </option>
              ))}
          </Select>
        </FormControl>
        <PhotoDropzone onSelect={upload} pending={pending} />
        {uploadedPhotos.length > 0 ? (
          <Stack spacing={3} pt={2}>
            {uploadedPhotos.map((photo) => (
              <HStack key={photo.id} spacing={3}>
                <PhotoPreview photo={photo} api={api} size="56px" />
                <Box>
                  <Text fontWeight="semibold">{photo.title}</Text>
                  <HStack spacing={2} mt={1}>
                    <PhotoStatus status={photo.status} />
                    <ProcessingStatus status={photo.processingStatus} />
                  </HStack>
                </Box>
              </HStack>
            ))}
          </Stack>
        ) : null}
      </Stack>
    </Page>
  );
};

const PhotoDetailPage = ({ api }) => {
  const { id } = useParams();
  const { collections, collectionError } = usePhotoCollections(api);
  const [photo, setPhoto] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const toast = useToast();
  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setPhoto(await api.getPhoto(id));
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api, id]);
  useEffect(() => {
    load();
  }, [load]);
  const handleMutationError = async (reason) => {
    if (
      reason instanceof GalleryApiError &&
      reason.code === "version_conflict"
    ) {
      toast({
        title: "Photo changed elsewhere",
        description: "The latest server version has been loaded.",
        status: "warning",
        duration: 5000,
        isClosable: true,
      });
      await load();
    } else {
      setError(reason);
    }
  };
  const save = async (form) => {
    setPending(true);
    setError(null);
    try {
      setPhoto(await api.updatePhoto(id, form));
      toast({
        title: "Photo saved",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
    } catch (reason) {
      await handleMutationError(reason);
    } finally {
      setPending(false);
    }
  };
  const transition = async (action) => {
    setPending(true);
    setError(null);
    try {
      const updated =
        action === "publish"
          ? await api.publishPhoto(id, photo.version)
          : action === "archive"
            ? await api.archivePhoto(id, photo.version)
            : await api.restorePhoto(id, photo.version);
      setPhoto(updated);
      toast({
        title:
          action === "publish"
            ? "Photo published"
            : action === "archive"
              ? "Photo archived"
              : "Photo restored as draft",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
    } catch (reason) {
      await handleMutationError(reason);
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
  if (!photo)
    return (
      <Page>
        <ErrorNotice error={error || collectionError} />
      </Page>
    );
  const formValue = {
    ...photo,
    tags: photo.tags || [],
    tagsText: (photo.tags || []).join(", "),
  };
  return (
    <Page>
      <Flex
        justify="space-between"
        align={{ base: "flex-start", sm: "center" }}
        gap={4}
        mb={7}
        direction={{ base: "column", sm: "row" }}
      >
        <HStack spacing={3}>
          <IconButton
            as={RouterLink}
            to="/photos"
            aria-label="Back to photos"
            icon={<ArrowBackIcon />}
            variant="solid"
            colorScheme="yellow"
          />
          <Box>
            <Heading size="lg">Edit photo</Heading>
            <HStack mt={1} spacing={3}>
              <PhotoStatus status={photo.status} />
              <ProcessingStatus status={photo.processingStatus} />
              <Text color="whiteAlpha.600">Version {photo.version}</Text>
            </HStack>
          </Box>
        </HStack>
        <HStack>
          {photo.status === "draft" ? (
            <Button
              colorScheme="green"
              onClick={() => transition("publish")}
              isLoading={pending}
              isDisabled={
                photo.processingStatus &&
                photo.processingStatus !== "ready" &&
                photo.processingStatus !== "not_required"
              }
            >
              Publish
            </Button>
          ) : null}
          {photo.status === "draft" || photo.status === "published" ? (
            <Button
              colorScheme="orange"
              onClick={() => transition("archive")}
              isLoading={pending}
            >
              Archive
            </Button>
          ) : null}
          {photo.status === "archived" ? (
            <Button
              colorScheme="blue"
              onClick={() => transition("restore")}
              isLoading={pending}
            >
              Restore
            </Button>
          ) : null}
        </HStack>
      </Flex>
      <ErrorNotice error={error || collectionError} />
      <PhotoForm
        initialValue={formValue}
        collections={collections}
        onSubmit={save}
        pending={pending}
        submitLabel="Save changes"
        readOnly={photo.status === "archived"}
        api={api}
      />
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
      <Route path="/photos/new" element={<PhotoCreatePage api={api} />} />
      <Route path="/photos/:id" element={<PhotoDetailPage api={api} />} />
      <Route path="*" element={<Navigate to="/collections" replace />} />
    </Routes>
  </Box>
);
