/*
Copyright (c) 2018 Ady Levy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
import React, { Component, Fragment } from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import compose from 'recompose/compose';
import { withStyles } from '@material-ui/core/styles';
import { fade } from '@material-ui/core/styles/colorManipulator';
import ActionDelete from '@material-ui/icons/Delete';
import classnames from 'classnames';
import { translate, crudDelete, startUndoable } from 'ra-core';
import IconCancel from '@material-ui/icons/Cancel';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';

import { SimpleForm, Button, DeleteButton, TextInput, LongTextInput, required } from 'react-admin';

const styles = (theme) => ({
  deleteButton: {
    color: theme.palette.error.main,
    '&:hover': {
      backgroundColor: fade(theme.palette.error.main, 0.12),
      // Reset on mouse devices
      '@media (hover: none)': {
        backgroundColor: 'transparent'
      }
    }
  }
});

class DeleteButtonWithConfirmation extends Component {
  state = {
    showDialog: false
  };

  handleClick = () => {
    this.setState({ showDialog: true });
  };

  handleCloseClick = () => {
    this.setState({ showDialog: false });
  };
  
  handleDelete = (event) => {
    event.preventDefault();
    this.setState({ showDialog: false });
    const { dispatchCrudDelete, startUndoable, resource, record, basePath, redirect, undoable } = this.props;
    if (undoable) {
      startUndoable(crudDelete(resource, record.id, record, basePath, redirect));
    } else {
      dispatchCrudDelete(resource, record.id, record, basePath, redirect);
    }
  };

  render() {
    const { showDialog } = this.state;
    const { label = 'ra.action.delete', classes = {}, className } = this.props;
    return (
      <Fragment>
        <Button onClick={this.handleClick} label={label} className={classnames('ra-delete-button', classes.deleteButton, className)} key="button">
          <ActionDelete />
        </Button>
        <Dialog fullWidth open={showDialog} onClose={this.handleCloseClick} aria-label="Are you sure?">
          <DialogTitle>Are you sure you want to delete this entity?</DialogTitle>
          <DialogContent>
            <div>
              Be patient, it may take a while
            </div>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.handleDelete} label={label} className={classnames('ra-delete-button', classes.deleteButton, className)} key="button">
              <ActionDelete />
            </Button>
            <Button label="ra.action.cancel" onClick={this.handleCloseClick}>
              <IconCancel />
            </Button>
          </DialogActions>
        </Dialog>
      </Fragment>
    );
  }
}

DeleteButtonWithConfirmation.propTypes = {
  basePath: PropTypes.string,
  classes: PropTypes.object,
  className: PropTypes.string,
  dispatchCrudDelete: PropTypes.func.isRequired,
  label: PropTypes.string,
  record: PropTypes.object,
  redirect: PropTypes.oneOfType([PropTypes.string, PropTypes.bool, PropTypes.func]),
  resource: PropTypes.string.isRequired,
  startUndoable: PropTypes.func,
  translate: PropTypes.func,
  undoable: PropTypes.bool
};

DeleteButtonWithConfirmation.defaultProps = {
  redirect: 'list',
  undoable: true
};

export default compose(
  connect(
    null,
    { startUndoable, dispatchCrudDelete: crudDelete }
  ),
  translate,
  withStyles(styles)
)(DeleteButtonWithConfirmation);
