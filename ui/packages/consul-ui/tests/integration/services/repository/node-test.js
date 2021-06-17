import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';
import { get } from '@ember/object';
const NAME = 'node';
module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const id = 'token-name';
  const now = new Date().getTime();
  const nspace = 'default';
  skip('findByDatacenter returns the correct data for list endpoint', function (assert) {
    get(this.owner.lookup(`service:repository/${NAME}`), 'store').serializerFor(NAME).timestamp =
      function () {
        return now;
      };
    return repo(
      'Node',
      'findAllByDatacenter',
      this.owner.lookup(`service:repository/${NAME}`),
      function retrieveStub(stub) {
        return stub(`/v1/internal/ui/nodes?dc=${dc}`, {
          CONSUL_NODE_COUNT: '100',
        });
      },
      function performTest(service) {
        return service.findAllByDatacenter({ dc });
      },
      function performAssertion(actual, expected) {
        actual.forEach((item) => {
          assert.equal(item.uid, `["${nspace}","${dc}","${item.ID}"]`);
          assert.equal(item.Datacenter, dc);
        });
      }
    );
  });
  skip('findBySlug returns the correct data for item endpoint', function (assert) {
    return repo(
      'Node',
      'findBySlug',
      this.owner.lookup(`service:repository/${NAME}`),
      function (stub) {
        return stub(`/v1/internal/ui/node/${id}?dc=${dc}`);
      },
      function (service) {
        return service.findBySlug({ id, dc });
      },
      function (actual, expected) {
        assert.equal(actual.uid, `["${nspace}","${dc}","${actual.ID}"]`);
        assert.equal(actual.Datacenter, dc);
      }
    );
  });
});
