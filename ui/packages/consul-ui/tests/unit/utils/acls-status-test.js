import { module, test } from 'qunit';
import sinon from 'sinon';
import aclsStatus from 'consul-ui/utils/acls-status';

module('Unit | Utility | acls status', function () {
  test('it rejects and nothing is enabled or authorized', function (assert) {
    const isValidServerError = sinon.stub().returns(false);
    const status = aclsStatus(isValidServerError);
    [
      sinon.stub().rejects(),
      sinon.stub().rejects({ errors: [] }),
      sinon.stub().rejects({ errors: [{ status: '404' }] }),
    ].forEach(function (reject) {
      const actual = status({
        response: reject(),
      });
      assert.rejects(actual.response);
      ['isAuthorized', 'isEnabled'].forEach(function (prop) {
        actual[prop].then(function (actual) {
          assert.notOk(actual);
        });
      });
    });
  });
  test('with a 401 it resolves with an empty array and nothing is enabled or authorized', async function (assert) {
    assert.expect(3);
    const isValidServerError = sinon.stub().returns(false);
    const status = aclsStatus(isValidServerError);
    const actual = status({
      response: sinon.stub().rejects({ errors: [{ status: '401' }] })(),
    });

    assert.deepEqual(await actual.response, []);
    ['isAuthorized', 'isEnabled'].forEach(function (prop) {
      actual[prop].then(function (actual) {
        assert.notOk(actual, `not ${prop}`);
      });
    });
  });
  test("with a 403 it resolves with an empty array and it's enabled but not authorized", async function (assert) {
    assert.expect(3);
    const isValidServerError = sinon.stub().returns(false);
    const status = aclsStatus(isValidServerError);
    const actual = status({
      response: sinon.stub().rejects({ errors: [{ status: '403' }] })(),
    });
    assert.deepEqual(await actual.response, []);
    actual.isEnabled.then(function (actual) {
      assert.ok(actual);
    });
    actual.isAuthorized.then(function (actual) {
      assert.notOk(actual);
    });
  });
  test("with a 500 (but not a 'valid' error) it rejects and nothing is enabled or authorized", function (assert) {
    assert.expect(3);
    const done = assert.async(2);
    const isValidServerError = sinon.stub().returns(false);
    const status = aclsStatus(isValidServerError);
    const actual = status({
      response: sinon.stub().rejects({ errors: [{ status: '500' }] })(),
    });
    assert.rejects(actual.response);
    ['isAuthorized', 'isEnabled'].forEach(function (prop) {
      actual[prop].then(function (actual) {
        assert.notOk(actual);
        done();
      });
    });
  });
  test("with a 500 and a 'valid' error, it resolves with an empty array and it's enabled but not authorized", function (assert) {
    assert.expect(3);
    const done = assert.async(3);
    const isValidServerError = sinon.stub().returns(true);
    const status = aclsStatus(isValidServerError);
    const actual = status({
      response: sinon.stub().rejects({ errors: [{ status: '500' }] })(),
    });
    actual.response.then(function (actual) {
      assert.deepEqual(actual, []);
      done();
    });
    actual.isEnabled.then(function (actual) {
      assert.ok(actual);
      done();
    });
    actual.isAuthorized.then(function (actual) {
      assert.notOk(actual);
      done();
    });
  });
});
