// in myRestProvider.js
import {
    GET_LIST,
    GET_ONE,
    CREATE,
    DELETE,
} from 'react-admin';
const apiUrl = window.location.protocol+'//'+window.location.hostname+(window.location.port ? ':'+window.location.port: '');

/**
 * Maps react-admin queries to my REST API
 *
 * @param {string} type Request type, e.g GET_LIST
 * @param {string} resource Resource name, e.g. "posts"
 * @param {Object} payload Request parameters. Depends on the request type
 * @returns {Promise} the Promise for a data response
 */
export default (type, resource, params) => {
    let url = '';
    var clustername = '';
    const options = {
        headers : new Headers({
            Accept: 'application/json',
        }),
    };
    switch (type) {
        case GET_LIST: {
            url = `${apiUrl}/api/v1/${resource}`;

            break;
        }
        case GET_ONE:
            url = `${apiUrl}/api/v1/${resource}/${params.id}/info`;
            break;
        case CREATE:
            url = `${apiUrl}/api/v1/${resource}/new`;
            options.method = 'POST';
            options.headers.append('Content-Type', 'application/x-www-form-urlencoded');
        
            options.body = "name="+params.data.name;
            clustername= params.data.name
            break;
        case DELETE:
            url = `${apiUrl}/api/v1/${resource}/${params.id}`;
            options.method = 'DELETE';
            break;
   
        default:
            throw new Error(`Unsupported Data Provider request type ${type}`);
    }

    return fetch(url, options)
        .then(res => res.json())
        .then(json => {
            const obj =                 new Array();
            const cl = {}
            var i = 1
            switch (type) {
                case GET_ONE:
                json.id = json.Name
                return {
                    data: json,
                    total: 1,
                };
                case GET_LIST:
             
                if (json.AvailableClusters != null) {
 for (const key of json.AvailableClusters) {
    obj.push( { id : key, name: key});
    i++
}
                }
                    return {
                      
                        data: obj || [],
                      total: obj.length
                    };
                case CREATE:
            
                    return {
                      
                        data: { id : clustername, name: clustername} ,
                      total: 1
                    };
                case DELETE:
                if (json.AvailableClusters != null) {
 for (const key of json.AvailableClusters) {
    obj.push( { id : key, name: key});
    i++
}
                }
                    return {
                      
                        data: obj || [],
                      total: obj.length
                    };
                default:
                    return { data: json };
            }
        });
};