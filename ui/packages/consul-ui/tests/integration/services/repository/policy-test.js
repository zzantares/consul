import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import { get } from '@ember/object';
import repo from 'consul-ui/tests/helpers/repo';
const NAME = 'policy';

module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  skip('translate returns the correct data for the translate endpoint');
  const now = new Date().getTime();
  const dc = 'dc-1';
  const id = 'policy-name';
  const undefinedNspace = 'default';
  [undefinedNspace, 'team-1', undefined].forEach((nspace) => {
    skip(`findByDatacenter returns the correct data for list endpoint when nspace is ${nspace}`, function (assert) {
      get(this.owner.lookup(`service:repository/${NAME}`), 'store').serializerFor(NAME).timestamp =
        function () {
          return now;
        };
      return repo(
        'Policy',
        'findAllByDatacenter',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveStub(stub) {
          return stub(
            `/v1/acl/policies?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`,
            {
              CONSUL_POLICY_COUNT: '100',
            }
          );
        },
        function performTest(service) {
          return service.findAllByDatacenter({ dc, ns: nspace || undefinedNspace });
        },
        function performAssertion(actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              return payload.map((item) =>
                Object.assign({}, item, {
                  SyncTime: now,
                  Datacenter: dc,
                  Namespace: item.Namespace || undefinedNspace,
                  uid: `["${item.Namespace || undefinedNspace}","${dc}","${item.ID}"]`,
                })
              );
            })
          );
        }
      );
    });
    skip(`findBySlug returns the correct data for item endpoint when the nspace is ${nspace}`, function (assert) {
      return repo(
        'Policy',
        'findBySlug',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveStub(stub) {
          return stub(
            `/v1/acl/policy/${id}?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`
          );
        },
        function performTest(service) {
          return service.findBySlug({ id, dc, ns: nspace || undefinedNspace });
        },
        function performAssertion(actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              const item = payload;
              return Object.assign({}, item, {
                Datacenter: dc,
                Namespace: item.Namespace || undefinedNspace,
                uid: `["${item.Namespace || undefinedNspace}","${dc}","${item.ID}"]`,
                meta: {
                  cacheControl: undefined,
                  cursor: undefined,
                  dc: dc,
                  nspace: item.Namespace || undefinedNspace,
                },
              });
            })
          );
        }
      );
    });
  });
});
