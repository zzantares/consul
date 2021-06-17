import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';

const NAME = 'auth-method';

module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const id = 'auth-method-name';
  const undefinedNspace = 'default';
  [undefinedNspace, 'team-1', undefined].forEach((nspace) => {
    skip(`findAllByDatacenter returns the correct data for list endpoint when nspace is ${nspace}`, function (assert) {
      return repo(
        'auth-method',
        'findAllByDatacenter',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveStub(stub) {
          return stub(
            `/v1/acl/auth-methods?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`,
            {
              CONSUL_AUTH_METHOD_COUNT: '3',
            }
          );
        },
        function performTest(service) {
          return service.findAllByDatacenter({
            dc: dc,
            nspace: nspace || undefinedNspace,
          });
        },
        function performAssertion(actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              return payload.map(function (item) {
                return Object.assign({}, item, {
                  Datacenter: dc,
                  Namespace: item.Namespace || undefinedNspace,
                  uid: `["${item.Namespace || undefinedNspace}","${dc}","${item.Name}"]`,
                });
              });
            })
          );
        }
      );
    });
  });
});
