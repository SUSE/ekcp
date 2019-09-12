import { ShowController } from 'ra-core';
import React from 'react';
import {
    ArrayField,
    Datagrid,
Link,
    ShowView,
    Tab,
    TabbedShowLayout,
    TextField,
} from 'react-admin'; // eslint-disable-line import/no-unresolved
import ClusterTitle from './ClusterTitle';
import Button from '@material-ui/core/Button';
import DeleteButtonWithConfirmation from './DeleteButtonWithConfirmation';



const ClusterShow = props => (
    <ShowController title={<ClusterTitle />}  {...props}>
        {controllerProps => (
            <ShowView {...props} {...controllerProps}>
             <TabbedShowLayout>
                    <Tab label="Summary">
                <TextField source="Name" />
                <TextField source="ClusterIP" />
                <DeleteButtonWithConfirmation record={record => recontrollerProps.record.name} resource={"cluster"} undoable={false} />
                </Tab>
                <Tab label="Routes">
                <ArrayField source="Routes">
                            <Datagrid>
                            <TextField source="Domain" />
                                <TextField source="Host" />
                                <TextField source="Port" />
                                <TextField label="SSL" source="TLSPort" />
                            </Datagrid>
                        </ArrayField>
                </Tab>
                <Tab label="Proxy">
                <TextField labal="Proxy URL" source="ProxyURL" />
                <TextField source="Kubeconfig" />

                <a target="_blank" href={controllerProps.record.Kubeconfig}>
                    <Button >
                        <p>Kubeconfig</p>
                    </Button>
                 </a>
                </Tab>
                { controllerProps.record.Federated &&
                <Tab label="Federation">
                <TextField source="InstanceEndpoint" />
                <a target="_blank" href={`${controllerProps.record.InstanceEndpoint}/ui`}>
                    <Button >
                        <p>Go to UI</p>
                    </Button>
                 </a>
                </Tab>
                }
                </TabbedShowLayout> 

            </ShowView>
        )}

    </ShowController>

);

export default ClusterShow;
