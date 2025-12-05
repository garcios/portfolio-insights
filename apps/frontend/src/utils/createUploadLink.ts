import { ApolloLink, Observable } from '@apollo/client';

export const createUploadLink = (options: { uri: string }) => {
    return new ApolloLink((operation) => {
        return new Observable((observer) => {
            const { variables, operationName } = operation;



            // We need to traverse the original variables to find files because JSON.stringify/parse loses File objects
            // But wait, we can't easily traverse and replace in one go if we want to keep the structure for JSON.
            // Let's do a simpler approach: traverse the variables, find files, and build the map.
            // Then create the operations object with nulls where files were.

            // Actually, let's just implement a simple traversal that modifies a clone or builds the map.
            // Since we know exactly where the file is in our specific use case (variables.file), 
            // we could hardcode it, but a generic solution is better.

            // Simplified generic approach:
            const map: Record<string, string[]> = {};
            const uploads: File[] = [];

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const scanAndReplace = (obj: any, path: string): any => {
                if (obj instanceof File) {
                    const index = uploads.length;
                    uploads.push(obj);
                    map[index] = [`variables.${path}`];
                    return null;
                }

                if (Array.isArray(obj)) {
                    return obj.map((item, i) => scanAndReplace(item, `${path}.${i}`));
                }

                if (obj && typeof obj === 'object') {
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    const result: any = {};
                    for (const key in obj) {
                        // Handle the root case where path is empty
                        const nextPath = path ? `${path}.${key}` : key;
                        result[key] = scanAndReplace(obj[key], nextPath);
                    }
                    return result;
                }

                return obj;
            };

            const cleanVariables = scanAndReplace(variables, '');

            // If no files found, use standard JSON fetch (or just let this handle it too, but standard is usually better for non-upload)
            // However, to keep it simple, if no files, we can just send JSON.
            if (uploads.length === 0) {
                fetch(options.uri, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        ...operation.getContext().headers,
                    },
                    body: JSON.stringify({
                        query: operation.query.loc?.source.body,
                        variables,
                        operationName,
                    }),
                })
                    .then(response => response.json())
                    .then(result => {
                        observer.next(result);
                        observer.complete();
                    })
                    .catch(error => {
                        observer.error(error);
                    });
                return;
            }

            // Prepare FormData for upload
            const formData = new FormData();

            const operations = {
                query: operation.query.loc?.source.body,
                variables: cleanVariables,
                operationName,
            };

            formData.append('operations', JSON.stringify(operations));
            formData.append('map', JSON.stringify(map));

            uploads.forEach((file, index) => {
                formData.append(`${index}`, file);
            });

            // Perform fetch
            fetch(options.uri, {
                method: 'POST',
                headers: {
                    // Content-Type is automatically set by browser for FormData (multipart/form-data)
                    // Do NOT set it manually, or boundary will be missing
                    ...operation.getContext().headers,
                },
                body: formData,
            })
                .then(response => response.json())
                .then(result => {
                    observer.next(result);
                    observer.complete();
                })
                .catch(error => {
                    observer.error(error);
                });
        });
    });
};
