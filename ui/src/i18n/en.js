import englishMessages from 'ra-language-english';
import treeEnglishMessages from 'ra-tree-language-english';
import { mergeTranslations } from 'react-admin';

export const messages = {
    simple: {
        action: {
            close: 'Close',
            resetViews: 'Reset views',
        },
        'create-cluster': 'New cluster',
    },
    ...mergeTranslations(englishMessages, treeEnglishMessages),
    cluster: {
        list: {
            search: 'Search',
        },
        form: {
            summary: 'Summary',
            body: 'Body',
            miscellaneous: 'Miscellaneous',
            comments: 'Comments',
        },
        edit: {
            title: 'Cluster "%{title}"',
        },
        action: {
            create: 'Create',
        },
    },
  
};

export default messages;
