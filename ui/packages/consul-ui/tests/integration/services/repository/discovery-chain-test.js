import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';

const NAME = 'discovery-chain';
module(`Integration | Repository | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const id = 'slug';
  skip('findBySlug returns the correct data for item endpoint', function (assert) {
    return repo(
      'DiscoveryChain',
      'findBySlug',
      this.owner.lookup(`service:repository/${NAME}`),
      function retrieveStub(stub) {
        return stub(`/v1/discovery-chain/${id}?dc=${dc}`, {
          CONSUL_DISCOVERY_CHAIN_COUNT: 1,
        });
      },
      function performTest(service) {
        return service.findBySlug({ id, dc });
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
