export default (collection, clickable, text) => () => {
  return collection('.consul-auth-method-list [data-test-list-row]', {
    authMethod: clickable('a'),
    methodName: text('[data-test-method-name]'),
    name: text('[data-test-name]'),
    type: text('[data-test-type]'),
  });
};
