const Yadda = require('yadda');
const English = Yadda.localisation.English;
const debug = require('./debug/index.js');

const scenario = English.library();
const currentURL = async (step) => {
  let url = await step.page.url()
  return url.replace(step.root, '');
}
debug(scenario, () => {}, currentURL);
module.exports = scenario.
  given(
    '$number $model model[s]? with the value "$value"',
    async function(number, model, value) {
      try {
        switch(model) {
          case 'datacenter':
            this.hcl.push(`${model} = "${value}"`);
            break;
        }
      } catch(e) {
        console.log(e);
      }
    }
  ).when(
    'I visit the $name page',
    async function(name) {
      try {
        const args = [
            'agent',
            '-dev',
            `-ui-content-path`,
            `${this.contentPath}`
        ].concat(
          this.hcl.reduce((prev, item) => prev.concat(['-hcl', `${item}`]), [])
        );
        await this.consul(args);
        await this.page.goto(`${this.root}`);
        await new Promise(
          (resolve) => {
            setTimeout(resolve, 1000)
          }
        );
      } catch(e) {
        console.log(e);
      }
    }
  ).then(
    'the url should be $url',
    async function(url) {
      const assert = {
        equal: (actual, expected, message) => {
          this.t.equal(actual, expected, message);
        }
      };
      // TODO: nice! $url should be wrapped in ""
      if (url === "''") {
        url = '';
      }
      let current = await currentURL(this) || '';
      assert.equal(current, url, `Expected the url to be ${url} was ${current}`);
    }
  );
