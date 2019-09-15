import BookIcon from '@material-ui/icons/Book';
import { withStyles } from '@material-ui/core/styles';
import React, { Children, Fragment, cloneElement } from 'react';
import {
    BulkDeleteButton,
    Datagrid,
    List,
    Responsive,
    ShowButton,
    SimpleList,
    TextField,
} from 'react-admin'; // eslint-disable-line import/no-unresolved

import ResetViewsButton from './ResetViewsButton';
export const FederationIcon = BookIcon;
// See : https://github.com/marmelab/react-admin/blob/master/docs/Theming.md

const styles = theme => ({
    title: {
        maxWidth: '20em',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
    },
    hiddenOnSmallScreens: {
        [theme.breakpoints.down('md')]: {
            display: 'none',
        },
    },
    publishedAt: { fontStyle: 'italic' },
});


const FederationListActionToolbar = withStyles({
    toolbar: {
        alignItems: 'center',
        display: 'flex',
    },
})(({ classes, children, ...props }) => (
    <div className={classes.toolbar}>
        {Children.map(children, button => cloneElement(button, props))}
        
    </div>
));

const rowClick = (id, basePath, record) => {
   

    return 'show';
};

const FederationPanel = ({ id, record, resource }) => (
    <div dangerouslySetInnerHTML={{ __html: record.body }} />
);

const FederationList = withStyles(styles)(({ classes, ...props }) => (
    <List
        {...props}
        bulkActionButtons={false}
    >
        <Responsive
            small={
                <SimpleList
                    primaryText={record => record.Endpoint}
                    secondaryText={record =>
                        record.Endpoint
                    }
                />
            }
            medium={
            // <Datagrid rowClick={rowClick} expand={<FederationPanel />}>
                <Datagrid>
                    <TextField label="Federation Endpoint" source="Endpoint"  cellClassName={classes.title} />
                    
                    <FederationListActionToolbar>
                        <ShowButton />
            
                    </FederationListActionToolbar>
                </Datagrid>
            }
        />
    </List>
));

export default FederationList;
