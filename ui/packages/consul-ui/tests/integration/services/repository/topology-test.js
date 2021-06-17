import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';

module('Integration | Repository | topology', function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const id = 'slug';
  const kind = '';
  skip('findBySlug returns the correct data for item endpoint', function (assert) {
    return repo(
      'Service',
      'findBySlug',
      this.owner.lookup('service:repository/topology'),
      function retrieveStub(stub) {
        return stub(`/v1/internal/ui/service-topology/${id}?dc=${dc}&${kind}`, {
          CONSUL_DISCOVERY_CHAIN_COUNT: 1,
        });
      },
      function performTest(service) {
        return service.findBySlug({ id, kind, dc });
      },
      function performAssertion(actual, expected) {
        const result = expected(function (payload) {
          return Object.assign(
            {},
            {
              Datacenter: dc,
              uid: `["default","${dc}","${id}"]`,
              meta: {
                cacheControl: undefined,
                cursor: undefined,
              },
            },
            payload
          );
        });
        assert.equal(actual.Datacenter, result.Datacenter);
        assert.equal(actual.uid, result.uid);
      }
    );
  });
});
