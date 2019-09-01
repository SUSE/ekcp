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
export const ClusterIcon = BookIcon;


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


const ClusterListActionToolbar = withStyles({
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

const ClusterPanel = ({ id, record, resource }) => (
    <div dangerouslySetInnerHTML={{ __html: record.body }} />
);

const ClusterList = withStyles(styles)(({ classes, ...props }) => (
    <List
        {...props}
        bulkActionButtons={false}
    >
        <Responsive
            small={
                <SimpleList
                    primaryText={record => record.name}
                    secondaryText={record =>
                        record.name
                    }
                />
            }
            medium={
            // <Datagrid rowClick={rowClick} expand={<ClusterPanel />}>
                <Datagrid>
                    <TextField source="id" />
                    <TextField source="name" cellClassName={classes.title} />
               
                    <ClusterListActionToolbar>
                        <ShowButton />
            
                    </ClusterListActionToolbar>
                </Datagrid>
            }
        />
    </List>
));

export default ClusterList;
