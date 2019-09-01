/* eslint react/jsx-key: off */
import React from 'react';
import { Admin, Resource } from 'react-admin'; // eslint-disable-line import/no-unresolved
import { render } from 'react-dom';

import ekcpdataProvider from './ekcpdataProvider';

import i18nProvider from './i18nProvider';
import clusters from './clusters';

render(
    <Admin
        dataProvider={ekcpdataProvider}
        i18nProvider={i18nProvider}
        title="EKCP Dashboard"
        locale="en"
  
    >
     <Resource name="cluster" {...clusters} />
    
    </Admin>,
    document.getElementById('root')
);
