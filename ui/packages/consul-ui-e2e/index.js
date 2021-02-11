const fs = require('fs');
const test = require('tape');
const Yadda = require('yadda');
const playwright = require('playwright');

const ua = playwright[process.env.PLAYWRIGHT_BROWSER || 'chromium'];
const { spawn } = require('child_process');
const library = require('./features/steps/index');

const parser = new Yadda.parsers.FeatureParser();
let page, browser, child;
const consul = (args) => {
  child = spawn('consul', args);
  return new Promise((resolve, reject) => {
    child.stdout.on(
      'data',
      (data) => {
        if(data.toString().indexOf('agent: Starting server:')) {
          setTimeout(
            resolve,
            500
          )
        }
      }
    )
    child.stderr.on(
      'data',
      (data) => {
        console.log(data.toString());
      }
    )
  });
}
const CONSUL_HTTP_ADDR = 'http://localhost:8500';
const headless = false;
const scenarios = {};
new Yadda.FeatureFileSearch('../consul-ui/tests/acceptance').each(function(file) {
  const text = fs.readFileSync(file, 'utf8');
  const feature = parser.parse(text);

  const yadda = Yadda.createInstance(library);
  if(feature.annotations.browsers) {
    feature.scenarios.forEach(scenario => {
      ['/consul', '/ui'].forEach(
        (contentPath) => {
          const root = `${CONSUL_HTTP_ADDR}${contentPath}`;
          scenarios[`${scenario.title} with the '${contentPath}' content-path`] = (context, done) => {
            try {
              return yadda.run(
                scenario.steps,
                {
                  ...context,
                  contentPath,
                  root
                },
                done
              )
            } catch(e) {
              console.log(e);
            }
		      };
        }
      )
    });

  }
});
Object.entries(scenarios).forEach(
  ([key, scenario]) => {
    test(
      key,
      async (t) => {
        // setup
        try {
          browser = await ua.launch({
            headless: headless
          });
          page = await browser.newPage();
        } catch(e) {
          console.log(e);
        }
        // test
        await new Promise(
          (resolve) => {
            scenario(
              {
                page: page,
                consul: consul,
                t: t,
                hcl: []
              },
              () => {
                t.end();resolve();
              }
            );
          }
        );
        // teardown
        if (!page.isClosed()) {
          await new Promise(
            (resolve) => {
              browser.on(
                'disconnected',
                () => {
                  browser = null;
                  page = null;
                  resolve();
                }
              )
              browser.close();
            }
          );
        }
        if(child) {
          await new Promise(
            (resolve) => {
              child.on(
                'exit',
                () => {
                  child = null;
                  resolve();
                }
              )
              child.kill('SIGKILL');
            }
          );
        }
      }
    );
  }
)


