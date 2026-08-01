import {
  Datagrid,
  List,
  NumberField,
  Show,
  SimpleShowLayout,
  TextField,
} from "react-admin";

export const CollectionList = () => (
  <List actions={false} exporter={false} perPage={25} sort={{ field: "order", order: "ASC" }}>
    <Datagrid bulkActionButtons={false} rowClick="show">
      <NumberField source="order" label="Order" />
      <TextField source="title" />
      <TextField source="slug" />
      <TextField source="coverPhotoId" label="Cover photo" />
    </Datagrid>
  </List>
);

export const CollectionShow = () => (
  <Show actions={false}>
    <SimpleShowLayout>
      <TextField source="title" />
      <TextField source="description" />
      <TextField source="slug" />
      <NumberField source="order" />
      <TextField source="coverPhotoId" label="Cover photo" />
    </SimpleShowLayout>
  </Show>
);
