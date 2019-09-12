import React from 'react';
import { translate } from 'react-admin';

export default translate(({ record, translate }) => (
    <span>
        {record ? translate('cluster.edit.title', { title: record.name }) : ''}
    </span>
));
