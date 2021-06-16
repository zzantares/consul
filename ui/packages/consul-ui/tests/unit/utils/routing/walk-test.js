import { module, test } from 'qunit';
import sinon from 'sinon';
import { walk } from 'consul-ui/utils/routing/walk';

module('Unit | Utility | routing/walk', function () {
  test('it walks down deep routes', function (assert) {
    const route = sinon.stub();
    const Router = {
      route: function (name, options, cb) {
        route();
        if (cb) {
          cb.apply(this, []);
        }
      },
    };
    walk.apply(Router, [
      {
        route: {
          _options: {
            path: '/:path',
          },
          next: {
            _options: {
              path: '/:path',
            },
            inside: {
              _options: {
                path: '/*path',
              },
            },
          },
        },
      },
    ]);
    assert.equal(route.callCount, 3);
  });
});
