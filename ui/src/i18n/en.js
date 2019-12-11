import englishMessages from 'ra-language-english';


export const messages = {
    simple: {
        action: {
            close: 'Close',
            resetViews: 'Reset views',
        },
        'create-cluster': 'New cluster',
    },
    ...englishMessages,
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
