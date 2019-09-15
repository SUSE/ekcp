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

const FederationCreateToolbar = props => (
    <Toolbar {...props}>
      
      
        <SaveButton
            label="cluster.action.create"
            redirect="show"
            submitOnEnter={false}
            variant="flat"
        />
       
       
    </Toolbar>
);

const FederationCreate = ({ permissions, ...props }) => (
    <Create {...props}>
        <SimpleForm
            toolbar={<FederationCreateToolbar />}
            validate={values => {
                const errors = {};
                ['Endpoint'].forEach(field => {
                    if (!values[field]) {
                        errors[field] = ['Required field'];
                    }
                });

            
                return errors;
            }}
        >
            <TextInput autoFocus source="Endpoint" 
                    defaultValue="http://endpointip:port"
                   />
            <FormDataConsumer>
                {({ formData, ...rest }) =>
                    formData.Endpoint 
                }
            </FormDataConsumer>
        </SimpleForm>
    </Create>
);

export default FederationCreate;
