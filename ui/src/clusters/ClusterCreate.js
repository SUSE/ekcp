import React, { Component } from 'react';
import { connect } from 'react-redux';

import {
   
    Create,
    FormDataConsumer,
    SaveButton,
    SimpleForm,
    TextInput,
    Toolbar,
    crudCreate,
} from 'react-admin'; // eslint-disable-line import/no-unresolved

const saveWithNote = (values, basePath, redirectTo) =>
    crudCreate('cluster', { ...values }, basePath, redirectTo);

class SaveWithNoteButtonComponent extends Component {
    handleClick = () => {
        const { basePath, handleSubmit, redirect, saveWithNote } = this.props;

        return handleSubmit(values => {
            saveWithNote(values, basePath, redirect);
        });
    };

    render() {
        const { handleSubmitWithRedirect, saveWithNote, ...props } = this.props;

        return (
            <SaveButton
                handleSubmitWithRedirect={this.handleClick}
                {...props}
            />
        );
    }
}

const SaveWithNoteButton = connect(
    undefined,
    { saveWithNote }
)(SaveWithNoteButtonComponent);

const ClusterCreateToolbar = props => (
    <Toolbar {...props}>
      
      
        <SaveButton
            label="cluster.action.create"
            redirect="show"
            submitOnEnter={false}
            variant="flat"
        />
       
       
    </Toolbar>
);

const ClusterCreate = ({ permissions, ...props }) => (
    <Create {...props}>
        <SimpleForm
            toolbar={<ClusterCreateToolbar />}
            validate={values => {
                const errors = {};
                ['name'].forEach(field => {
                    if (!values[field]) {
                        errors[field] = ['Required field'];
                    }
                });

            
                return errors;
            }}
        >
            <TextInput autoFocus source="name" 
                    defaultValue="testcluster"
                   />
            <FormDataConsumer>
                {({ formData, ...rest }) =>
                    formData.name 
                }
            </FormDataConsumer>
        </SimpleForm>
    </Create>
);

export default ClusterCreate;
