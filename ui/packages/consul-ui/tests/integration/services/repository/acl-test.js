import { module, skip } from 'qunit';
import { setupTest } from 'ember-qunit';
import repo from 'consul-ui/tests/helpers/repo';

const NAME = 'acl';
module(`Integration | Service | ${NAME}`, function (hooks) {
  setupTest(hooks);
  const dc = 'dc-1';
  const nspace = 'default';
  const id = 'token-name';
  skip('findByDatacenter returns the correct data for list endpoint', function (assert) {
    return repo(
      'Acl',
      'findAllByDatacenter',
      this.owner.lookup(`service:repository/${NAME}`),
      function retrieveStub(stub) {
        return stub(`/v1/acl/list?dc=${dc}`, {
          CONSUL_ACL_COUNT: '100',
        });
      },
      function performTest(service) {
        return service.findAllByDatacenter({ dc });
      },
      function performAssertion(actual, expected) {
        assert.deepEqual(
          actual,
          expected(function (payload) {
            return payload.map((item) =>
              Object.assign({}, item, {
                Datacenter: dc,
                // TODO: default isn't required here, once we've
                // refactored out our Serializer this can go
                uid: `["${nspace}","${dc}","${item.ID}"]`,
              })
            );
          })
        );
      }
    );
  });
  skip('findBySlug returns the correct data for item endpoint', function (assert) {
    return repo(
      'Acl',
      'findBySlug',
      this.owner.lookup(`service:repository/${NAME}`),
      function retrieveStub(stub) {
        return stub(`/v1/acl/info/${id}?dc=${dc}`);
      },
      function performTest(service) {
        return service.findBySlug({ id, dc });
      },
      function performAssertion(actual, expected) {
        assert.deepEqual(
          actual,
          expected(function (payload) {
            const item = payload[0];
            return Object.assign({}, item, {
              Datacenter: dc,
              // TODO: default isn't required here, once we've
              // refactored out our Serializer this can go
              uid: `["${nspace}","${dc}","${item.ID}"]`,
            });
          })
        );
      }
    );
  });
});
