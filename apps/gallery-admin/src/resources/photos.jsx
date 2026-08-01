import {
  BooleanField,
  Datagrid,
  List,
  NumberField,
  Show,
  SimpleShowLayout,
  TextField,
} from "react-admin";

export const PhotoList = () => (
  <List actions={false} exporter={false} perPage={25} sort={{ field: "order", order: "ASC" }}>
    <Datagrid bulkActionButtons={false} rowClick="show">
      <NumberField source="order" label="Order" />
      <TextField source="title" />
      <TextField source="collectionId" label="Collection" />
      <TextField source="year" />
      <TextField source="location" />
      <BooleanField source="featured" />
    </Datagrid>
  </List>
);

export const PhotoShow = () => (
  <Show actions={false}>
    <SimpleShowLayout>
      <TextField source="title" />
      <TextField source="description" />
      <TextField source="collectionId" label="Collection" />
      <TextField source="year" />
      <TextField source="location" />
      <NumberField source="width" />
      <NumberField source="height" />
      <BooleanField source="featured" />
      <NumberField source="order" />
      <TextField source="src" label="Current image URL" />
    </SimpleShowLayout>
  </Show>
);
