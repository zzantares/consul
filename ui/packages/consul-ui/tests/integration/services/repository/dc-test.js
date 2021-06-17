import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';
const NAME = 'dc';
module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  skip("findBySlug (doesn't interact with the API) but still needs an int test");
  skip('findAll returns the correct data for list endpoint', function (assert) {
    return repo(
      'Dc',
      'findAll',
      this.owner.lookup(`service:repository/${NAME}`),
      function retrieveStub(stub) {
        return stub(`/v1/catalog/datacenters`, {
          CONSUL_DATACENTER_COUNT: '100',
        });
      },
      function performTest(service) {
        return service.findAll();
      },
      function performAssertion(actual, expected) {
        assert.deepEqual(
          actual,
          expected(function (payload) {
            return payload.map((item, i) => ({ Name: item, Local: i === 0 ? true : false }));
          })
        );
      }
    );
  });
});
