import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';
import { get } from '@ember/object';
const NAME = 'session';

module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);

  const dc = 'dc-1';
  const id = 'node-name';
  const now = new Date().getTime();
  const undefinedNspace = 'default';
  [undefinedNspace, 'team-1', undefined].forEach((nspace) => {
    skip(`findByNode returns the correct data for list endpoint when the nspace is ${nspace}`, function (assert) {
      get(this.owner.lookup(`service:repository/${NAME}`), 'store').serializerFor(NAME).timestamp =
        function () {
          return now;
        };
      return repo(
        'Session',
        'findByNode',
        this.owner.lookup(`service:repository/${NAME}`),
        function retrieveStub(stub) {
          return stub(
            `/v1/session/node/${id}?dc=${dc}${
              typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``
            }`,
            {
              CONSUL_SESSION_COUNT: '100',
            }
          );
        },
        function performTest(service) {
          return service.findByNode({ id, dc, ns: nspace || undefinedNspace });
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
    skip(`findByKey returns the correct data for item endpoint when the nspace is ${nspace}`, function (assert) {
      return repo(
        'Session',
        'findByKey',
        this.owner.lookup(`service:repository/${NAME}`),
        function (stub) {
          return stub(
            `/v1/session/info/${id}?dc=${dc}${typeof nspace !== 'undefined' ? `&ns=${nspace}` : ``}`
          );
        },
        function (service) {
          return service.findByKey({ id, dc, ns: nspace || undefinedNspace });
        },
        function (actual, expected) {
          assert.deepEqual(
            actual,
            expected(function (payload) {
              const item = payload[0];
              return Object.assign({}, item, {
                Datacenter: dc,
                Namespace: item.Namespace || undefinedNspace,
                uid: `["${item.Namespace || undefinedNspace}","${dc}","${item.ID}"]`,
              });
            })
          );
        }
      );
    });
  });
});
