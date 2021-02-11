/* eslint no-console: "off" */
module.exports = function(scenario, assert, currentURL) {
  scenario
    .then('print the current url', async function(url) {
      const step = this;
      console.log(await currentURL(step));
    })
    .then('log the "$text"', async function(text) {
      console.log(text);
    })
    .then('pause for $milliseconds', async function(milliseconds) {
      return new Promise(function(resolve) {
        setTimeout(resolve, milliseconds);
      });
    })
    .then('ok', function() {
      const assert = this.t;
      assert.ok(true);
    });
}
