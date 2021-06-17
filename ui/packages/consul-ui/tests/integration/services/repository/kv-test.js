import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';
const NAME = 'kv';

module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const id = 'key-name';
  const undefinedNspace = 'default';
  [undefinedNspace, 'team-1', undefined].forEach((nspace) => {
    skip(`findAllBySlug returns the correct data for list endpoint when nspace is ${nspace}`, function (assert) {
      return repo(
        'Kv',
        'findAllBySlug',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveTest(stub) {
          return stub(
            `/v1/kv/${id}?keys&dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`,
            {
              CONSUL_KV_COUNT: '1',
            }
          );
        },
        function performTest(service) {
          return service.findAllBySlug({ id, dc, ns: nspace || undefinedNspace });
        },
        function performAssertion(actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              return payload.map((item) => {
                return {
                  Datacenter: dc,
                  Namespace: nspace || undefinedNspace,
                  uid: `["${nspace || undefinedNspace}","${dc}","${item}"]`,
                  Key: item,
                };
              });
            })
          );
        }
      );
    });
    skip(`findBySlug returns the correct data for item endpoint when nspace is ${nspace}`, function (assert) {
      return repo(
        'Kv',
        'findBySlug',
        this.owner.lookup(`service:repository/${NAME}`),
        function (stub) {
          return stub(
            `/v1/kv/${id}?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`
          );
        },
        function (service) {
          return service.findBySlug({ id, dc, ns: nspace || undefinedNspace });
        },
        function (actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              const item = payload[0];
              return Object.assign({}, item, {
                Datacenter: dc,
                Namespace: item.Namespace || undefinedNspace,
                uid: `["${item.Namespace || undefinedNspace}","${dc}","${item.Key}"]`,
              });
            })
          );
        }
      );
    });
  });
});
