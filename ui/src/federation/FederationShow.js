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
import FederationTitle from './FederationTitle';
import Button from '@material-ui/core/Button';
import DeleteButtonWithConfirmation from './DeleteButtonWithConfirmation';



const FederationShow = props => (
    <ShowController title={<FederationTitle />}  {...props}>
        {controllerProps => (
            <ShowView {...props} {...controllerProps}>
             <TabbedShowLayout>
                    <Tab label="Summary">
                <TextField source="Endpoint" />
                <DeleteButtonWithConfirmation record={record => controllerProps.record.Endpoint} resource={"federation"} undoable={false} />
                </Tab>
               
                </TabbedShowLayout> 

            </ShowView>
        )}

    </ShowController>

);

export default FederationShow;
