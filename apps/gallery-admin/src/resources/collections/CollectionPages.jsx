import { AddIcon, ArrowBackIcon, EditIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Flex,
  FormControl,
  FormLabel,
  Heading,
  HStack,
  IconButton,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  NumberInput,
  NumberInputField,
  Stack,
  Table,
  Tbody,
  Td,
  Text,
  Textarea,
  Th,
  Thead,
  Tr,
  useDisclosure,
  useToast,
} from "@chakra-ui/react";
import { useCallback, useEffect, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { ErrorNotice, EmptyState, LoadingState, Page, StatusBadge } from "../../components/ConsolePrimitives";
import { GalleryApiError } from "../../galleryDataProvider";

const collectionBlank = {
  title: "",
  slug: "",
  description: "",
  coverPhotoId: "",
  order: 1,
};

const CollectionForm = ({
  initialValue,
  onSubmit,
  pending,
  submitLabel,
  slugReadOnly = false,
  readOnly = false,
}) => {
  const [form, setForm] = useState(initialValue);

  // Replace stale local form state after a reload triggered by an optimistic
  // concurrency conflict, so an old version cannot be resubmitted.
  useEffect(() => setForm(initialValue), [initialValue]);

  const change = (field) => (event) =>
    setForm((current) => ({ ...current, [field]: event.target.value }));

  return (
    <Box
      as="form"
      onSubmit={(event) => {
        event.preventDefault();
        if (!readOnly) onSubmit(form);
      }}
    >
      <Stack spacing={5} maxW="680px">
        <FormControl isRequired>
          <FormLabel>Title</FormLabel>
          <Input value={form.title} onChange={change("title")} isReadOnly={readOnly} autoComplete="off" />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Slug</FormLabel>
          <Input value={form.slug} onChange={change("slug")} isReadOnly={slugReadOnly} autoComplete="off" />
          {slugReadOnly ? <Text mt={1} fontSize="sm" color="whiteAlpha.600">The URL identifier is immutable after creation.</Text> : null}
        </FormControl>
        <FormControl>
          <FormLabel>Description</FormLabel>
          <Textarea value={form.description} onChange={change("description")} isReadOnly={readOnly} rows={4} resize="vertical" />
        </FormControl>
        <FormControl>
          <FormLabel>Cover photo ID</FormLabel>
          <Input value={form.coverPhotoId} onChange={change("coverPhotoId")} isReadOnly={readOnly} autoComplete="off" />
        </FormControl>
        <FormControl isRequired>
          <FormLabel>Display order</FormLabel>
          <NumberInput min={1} value={form.order} isReadOnly={readOnly} onChange={(value) => setForm((current) => ({ ...current, order: value }))}>
            <NumberInputField />
          </NumberInput>
        </FormControl>
        {!readOnly ? (
          <HStack pt={2}>
            <Button type="submit" colorScheme="yellow" isLoading={pending} loadingText="Saving">{submitLabel}</Button>
            <Button as={RouterLink} to="/collections" variant="solid" colorScheme="green">Cancel</Button>
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
    if (await onDelete(confirmation)) close();
  };

  return (
    <>
      <Button colorScheme="red" onClick={onOpen} isLoading={pending}>Delete</Button>
      <Modal isOpen={isOpen} onClose={close} isCentered>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader color="red.500">Delete collection</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Text color="black">
              This permanently removes the archived collection metadata. Type <Text as="span" fontWeight="bold">{collection.slug}</Text> to confirm.
            </Text>
            <Input mt={4} value={confirmation} color={confirmed ? "green.500" : "red.500"} onChange={(event) => setConfirmation(event.target.value)} autoComplete="off" aria-label="Collection slug confirmation" />
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={close}>Cancel</Button>
            <Button colorScheme="red" onClick={remove} isDisabled={!confirmed || pending} isLoading={pending}>Permanently delete</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
};

export const CollectionsPage = ({ api }) => {
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

  useEffect(() => { load(); }, [load]);

  return (
    <Page>
      <Flex justify="space-between" align={{ base: "flex-start", sm: "center" }} gap={4} mb={7} direction={{ base: "column", sm: "row" }}>
        <Box>
          <Heading size="lg">Collections</Heading>
          <Text mt={2} color="whiteAlpha.700">Organize the gallery rooms and their public metadata.</Text>
        </Box>
        <Button leftIcon={<AddIcon />} colorScheme="yellow" as={RouterLink} to="/collections/new">New collection</Button>
      </Flex>
      <ErrorNotice error={error} />
      {loading ? <LoadingState /> : null}
      {!loading && !error && collections.length === 0 ? <EmptyState>No collections have been created yet.</EmptyState> : null}
      {!loading && !error && collections.length > 0 ? (
        <Box overflowX="auto" borderWidth="1px" borderColor="whiteAlpha.200" borderRadius="md">
          <Table variant="simple">
            <Thead><Tr><Th>Order</Th><Th>Title</Th><Th>Slug</Th><Th>Status</Th><Th>Cover photo</Th><Th aria-label="Actions" /></Tr></Thead>
            <Tbody>
              {collections.map((collection) => (
                <Tr key={collection.id} _hover={{ bg: "whiteAlpha.50" }}>
                  <Td>{collection.order}</Td>
                  <Td fontWeight="semibold">{collection.title}</Td>
                  <Td color="whiteAlpha.700">{collection.slug}</Td>
                  <Td><StatusBadge status={collection.status} /></Td>
                  <Td color="whiteAlpha.700">{collection.coverPhotoId || "-"}</Td>
                  <Td textAlign="right"><IconButton aria-label={`Edit ${collection.title}`} icon={<EditIcon />} size="sm" variant="solid" colorScheme="teal" onClick={() => navigate(`/collections/${collection.id}`)} /></Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </Box>
      ) : null}
    </Page>
  );
};

export const CollectionCreatePage = ({ api }) => {
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
        <IconButton as={RouterLink} to="/collections" aria-label="Back to collections" icon={<ArrowBackIcon />} variant="solid" colorScheme="yellow" />
        <Heading size="lg">New collection</Heading>
      </HStack>
      <ErrorNotice error={error} />
      <CollectionForm initialValue={collectionBlank} onSubmit={create} pending={pending} submitLabel="Create collection" />
    </Page>
  );
};

export const CollectionEditPage = ({ api }) => {
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

  useEffect(() => { load(); }, [load]);

  const handleMutationError = async (reason) => {
    if (reason instanceof GalleryApiError && reason.code === "version_conflict") {
      toast({ title: "Collection changed elsewhere", description: "The latest server version has been loaded.", status: "warning", duration: 5000, isClosable: true });
      await load();
    } else {
      setError(reason);
    }
  };

  const save = async (form) => {
    setPending(true);
    setError(null);
    try {
      setCollection(await api.updateCollection(id, form));
      toast({ title: "Collection saved", status: "success", duration: 3000, isClosable: true });
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
      const updated = action === "publish"
        ? await api.publishCollection(id, collection.version)
        : action === "archive"
          ? await api.archiveCollection(id, collection.version)
          : await api.restoreCollection(id, collection.version);
      setCollection(updated);
      toast({ title: action === "publish" ? "Collection published" : action === "archive" ? "Collection archived" : "Collection restored as draft", status: "success", duration: 3000, isClosable: true });
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
      toast({ title: "Collection permanently deleted", status: "success", duration: 3000, isClosable: true });
      navigate("/collections", { replace: true });
      return true;
    } catch (reason) {
      await handleMutationError(reason);
      return false;
    } finally {
      setPending(false);
    }
  };

  if (loading) return <Page><LoadingState /></Page>;
  // Keep loaded metadata and the back route available after a mutation fails.
  if (!collection) return <Page><ErrorNotice error={error} /></Page>;

  return (
    <Page>
      <Flex justify="space-between" align={{ base: "flex-start", sm: "center" }} gap={4} mb={7} direction={{ base: "column", sm: "row" }}>
        <HStack spacing={3}>
          <IconButton as={RouterLink} to="/collections" aria-label="Back to collections" icon={<ArrowBackIcon />} variant="solid" colorScheme="yellow" />
          <Box>
            <Heading size="lg">Edit collection</Heading>
            <HStack mt={1} spacing={3}><StatusBadge status={collection.status} /><Text color="whiteAlpha.600">Version {collection.version}</Text></HStack>
          </Box>
        </HStack>
        <HStack>
          {collection.status === "draft" ? <Button colorScheme="green" onClick={() => transition("publish")} isLoading={pending}>Publish</Button> : null}
          {collection.status === "draft" || collection.status === "published" ? <Button colorScheme="orange" onClick={() => transition("archive")} isLoading={pending}>Archive</Button> : null}
          {collection.status === "archived" ? <><Button colorScheme="blue" onClick={() => transition("restore")} isLoading={pending}>Restore</Button><DeleteCollectionControl collection={collection} pending={pending} onDelete={remove} /></> : null}
        </HStack>
      </Flex>
      <ErrorNotice error={error} />
      <CollectionForm initialValue={collection} onSubmit={save} pending={pending} submitLabel="Save changes" slugReadOnly readOnly={collection.status === "archived"} />
    </Page>
  );
};
