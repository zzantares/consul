import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import { get } from '@ember/object';
import repo from 'consul-ui/tests/helpers/repo';
import { createPolicies } from 'consul-ui/tests/helpers/normalizers';

const NAME = 'role';

module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const now = new Date().getTime();
  const dc = 'dc-1';
  const id = 'role-name';
  const undefinedNspace = 'default';
  [undefinedNspace, 'team-1', undefined].forEach((nspace) => {
    skip(`findByDatacenter returns the correct data for list endpoint when nspace is ${nspace}`, function (assert) {
      get(this.owner.lookup(`service:repository/${NAME}`), 'store').serializerFor(NAME).timestamp =
        function () {
          return now;
        };
      return repo(
        'Role',
        'findAllByDatacenter',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveStub(stub) {
          return stub(
            `/v1/acl/roles?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`,
            {
              CONSUL_ROLE_COUNT: '100',
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
                  Policies: createPolicies(item),
                })
              );
            })
          );
        }
      );
    });
    skip(`findBySlug returns the correct data for item endpoint when the nspace is ${nspace}`, function (assert) {
      return repo(
        'Role',
        'findBySlug',
        this.subject(),
        function retrieveStub(stub) {
          return stub(
            `/v1/acl/role/${id}?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`
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
                Policies: createPolicies(item),
              });
            })
          );
        }
      );
    });
  });
});
